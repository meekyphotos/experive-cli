package osm

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/meekyphotos/experive-cli/core/commands/pbf"
	"google.golang.org/protobuf/proto"
	"io"
	"sync"
	"time"
)

const (
	maxBlobHeaderSize = 64 * 1024

	initialBlobBufSize = 1 * 1024 * 1024

	// MaxBlobSize is maximum supported blob size.
	MaxBlobSize = 32 * 1024 * 1024
)

var (
	parseCapabilities = map[string]bool{
		"OsmSchema-V0.6": true,
		"DenseNodes":     true,
	}
)

type pair struct {
	i interface{}
	e error
}

type Decoder struct {
	r             io.Reader
	serializer    chan pair
	skipNodes     bool
	skipWays      bool
	skipRelations bool

	buf *bytes.Buffer

	// store header block
	header *Header
	// synchronize header deserialization
	headerOnce sync.Once

	// for data decoders
	inputs  []chan<- pair
	outputs []<-chan pair
}

// SetBufferSize sets initial size of decoding buffer. Default value is 1MB, you can set higher value
// (for example, MaxBlobSize) for (probably) faster decoding, or lower value for reduced memory consumption.
// Any value will produce valid results; buffer will grow automatically if required.
func (dec *Decoder) SetBufferSize(n int) {
	dec.buf = bytes.NewBuffer(make([]byte, 0, n))
}

func (dec *Decoder) Skip(nodes bool, ways bool, relation bool) {
	dec.skipNodes = nodes
	dec.skipWays = ways
	dec.skipRelations = relation
}

func NewDecoder(r io.Reader) *Decoder {
	d := &Decoder{
		r:          r,
		serializer: make(chan pair, 8000), // typical PrimitiveBlock contains 8k OSM entities
	}
	d.SetBufferSize(initialBlobBufSize)
	return d
}

// Header returns file header.
func (dec *Decoder) Header() (*Header, error) {
	// deserialize the file header
	return dec.header, dec.readOSMHeader()
}

// Start decoding process using n goroutines.
func (dec *Decoder) Start(n int, styler *GazetteerStyler) error {
	if n < 1 {
		n = 1
	}

	if err := dec.readOSMHeader(); err != nil {
		return err
	}

	// start data decoders
	for i := 0; i < n; i++ {
		input := make(chan pair)
		output := make(chan pair)
		go func() {
			dd := new(dataDecoder)
			dd.styler = styler
			dd.skipNodes = dec.skipNodes
			dd.skipWays = dec.skipWays
			dd.skipRelations = dec.skipRelations
			for p := range input {
				if p.e == nil {
					// send decoded objects or decoding error
					objects, err := dd.Decode(p.i.(*pbf.Blob))
					output <- pair{objects, err}
				} else {
					// send input error as is
					output <- pair{nil, p.e}
				}
			}
			close(output)
		}()

		dec.inputs = append(dec.inputs, input)
		dec.outputs = append(dec.outputs, output)
	}

	// start reading OSMData
	go func() {
		var inputIndex int
		for {
			input := dec.inputs[inputIndex]
			inputIndex = (inputIndex + 1) % n

			blobHeader, blob, err := dec.readFileBlock()
			if err == nil && blobHeader.GetType() != "OSMData" {
				err = fmt.Errorf("unexpected fileblock of type %s", blobHeader.GetType())
			}
			if err == nil {
				// send blob for decoding
				input <- pair{blob, nil}
			} else {
				// send input error as is
				input <- pair{nil, err}
				for _, input := range dec.inputs {
					close(input)
				}
				return
			}
		}
	}()

	go func() {
		var outputIndex int
		for {
			output := dec.outputs[outputIndex]
			outputIndex = (outputIndex + 1) % n

			p := <-output
			if p.i != nil {
				// send decoded objects one by one
				for _, o := range p.i.([]interface{}) {
					dec.serializer <- pair{o, nil}
				}
			}
			if p.e != nil {
				// send input or decoding error
				dec.serializer <- pair{nil, p.e}
				close(dec.serializer)
				return
			}
		}
	}()

	return nil
}

// Decode reads the next object from the input stream and returns either a
// pointer to Node, Way or Relation struct representing the underlying OpenStreetMap PBF
// data, or error encountered. The end of the input stream is reported by an io.EOF error.
//
// Decode is safe for parallel execution. Only first error encountered will be returned,
// subsequent invocations will return io.EOF.
func (dec *Decoder) Decode() (interface{}, error) {
	p, ok := <-dec.serializer
	if !ok {
		return nil, io.EOF
	}
	return p.i, p.e
}

func (dec *Decoder) readFileBlock() (*pbf.BlobHeader, *pbf.Blob, error) {
	blobHeaderSize, err := dec.readBlobHeaderSize()
	if err != nil {
		return nil, nil, err
	}

	blobHeader, err := dec.readBlobHeader(blobHeaderSize)
	if err != nil {
		return nil, nil, err
	}

	blob, err := dec.readBlob(blobHeader)
	if err != nil {
		return nil, nil, err
	}

	return blobHeader, blob, err
}

func (dec *Decoder) readBlobHeaderSize() (uint32, error) {
	dec.buf.Reset()
	if _, err := io.CopyN(dec.buf, dec.r, 4); err != nil {
		return 0, err
	}

	size := binary.BigEndian.Uint32(dec.buf.Bytes())

	if size >= maxBlobHeaderSize {
		return 0, errors.New("BlobHeader size >= 64Kb")
	}
	return size, nil
}

func (dec *Decoder) readBlobHeader(size uint32) (*pbf.BlobHeader, error) {
	dec.buf.Reset()
	if _, err := io.CopyN(dec.buf, dec.r, int64(size)); err != nil {
		return nil, err
	}

	blobHeader := new(pbf.BlobHeader)
	if err := proto.Unmarshal(dec.buf.Bytes(), blobHeader); err != nil {
		return nil, err
	}

	if blobHeader.GetDatasize() >= MaxBlobSize {
		return nil, errors.New("Blob size >= 32Mb")
	}
	return blobHeader, nil
}

func (dec *Decoder) readBlob(blobHeader *pbf.BlobHeader) (*pbf.Blob, error) {
	dec.buf.Reset()
	if _, err := io.CopyN(dec.buf, dec.r, int64(blobHeader.GetDatasize())); err != nil {
		return nil, err
	}

	blob := new(pbf.Blob)
	if err := proto.Unmarshal(dec.buf.Bytes(), blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func getData(blob *pbf.Blob) ([]byte, error) {
	switch blob.Data.(type) {
	case *pbf.Blob_Raw:
		return blob.GetRaw(), nil

	case *pbf.Blob_ZlibData:
		r, err := zlib.NewReader(bytes.NewReader(blob.GetZlibData()))
		if err != nil {
			return nil, err
		}
		buf := bytes.NewBuffer(make([]byte, 0, blob.GetRawSize()+bytes.MinRead))
		_, err = buf.ReadFrom(r)
		if err != nil {
			return nil, err
		}
		if buf.Len() != int(blob.GetRawSize()) {
			err = fmt.Errorf("raw blob data size %d but expected %d", buf.Len(), blob.GetRawSize())
			return nil, err
		}
		return buf.Bytes(), nil

	default:
		return nil, fmt.Errorf("unhandled blob data type %T", blob.Data)
	}
}

func (dec *Decoder) readOSMHeader() error {
	var err error
	dec.headerOnce.Do(func() {
		var blobHeader *pbf.BlobHeader
		var blob *pbf.Blob
		blobHeader, blob, err = dec.readFileBlock()
		if err == nil {
			if blobHeader.GetType() == "OSMHeader" {
				err = dec.decodeOSMHeader(blob)
			} else {
				err = fmt.Errorf("unexpected first fileblock of type %s", blobHeader.GetType())
			}
		}
	})

	return err
}

func (dec *Decoder) decodeOSMHeader(blob *pbf.Blob) error {
	data, err := getData(blob)
	if err != nil {
		return err
	}

	headerBlock := new(pbf.HeaderBlock)
	if err := proto.Unmarshal(data, headerBlock); err != nil {
		return err
	}

	// Check we have the parse capabilities
	requiredFeatures := headerBlock.GetRequiredFeatures()
	for _, feature := range requiredFeatures {
		if !parseCapabilities[feature] {
			return fmt.Errorf("parser does not have %s capability", feature)
		}
	}

	// Read properties to header struct
	header := &Header{
		RequiredFeatures:                 headerBlock.GetRequiredFeatures(),
		OptionalFeatures:                 headerBlock.GetOptionalFeatures(),
		WritingProgram:                   headerBlock.GetWritingprogram(),
		Source:                           headerBlock.GetSource(),
		OsmosisReplicationBaseUrl:        headerBlock.GetOsmosisReplicationBaseUrl(),
		OsmosisReplicationSequenceNumber: headerBlock.GetOsmosisReplicationSequenceNumber(),
	}

	// convert timestamp epoch seconds to golang time structure if it exists
	if headerBlock.OsmosisReplicationTimestamp != nil {
		header.OsmosisReplicationTimestamp = time.Unix(*headerBlock.OsmosisReplicationTimestamp, 0)
	}
	// read bounding box if it exists
	if headerBlock.Bbox != nil {
		// Units are always in nanodegree and do not obey granularity rules. See osmformat.proto
		header.BoundingBox = &BoundingBox{
			Left:   1e-9 * float64(*headerBlock.Bbox.Left),
			Right:  1e-9 * float64(*headerBlock.Bbox.Right),
			Bottom: 1e-9 * float64(*headerBlock.Bbox.Bottom),
			Top:    1e-9 * float64(*headerBlock.Bbox.Top),
		}
	}

	dec.header = header

	return nil
}
