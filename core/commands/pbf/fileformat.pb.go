// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.5.0
// source: fileformat.proto

package pbf

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Blob struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RawSize *int32 `protobuf:"varint,2,opt,name=raw_size,json=rawSize" json:"raw_size,omitempty"` // When compressed, the uncompressed size
	// Types that are assignable to Data:
	//	*Blob_Raw
	//	*Blob_ZlibData
	//	*Blob_LzmaData
	//	*Blob_OBSOLETEBzip2Data
	//	*Blob_Lz4Data
	//	*Blob_ZstdData
	Data isBlob_Data `protobuf_oneof:"data"`
}

func (x *Blob) Reset() {
	*x = Blob{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fileformat_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Blob) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Blob) ProtoMessage() {}

func (x *Blob) ProtoReflect() protoreflect.Message {
	mi := &file_fileformat_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Blob.ProtoReflect.Descriptor instead.
func (*Blob) Descriptor() ([]byte, []int) {
	return file_fileformat_proto_rawDescGZIP(), []int{0}
}

func (x *Blob) GetRawSize() int32 {
	if x != nil && x.RawSize != nil {
		return *x.RawSize
	}
	return 0
}

func (m *Blob) GetData() isBlob_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *Blob) GetRaw() []byte {
	if x, ok := x.GetData().(*Blob_Raw); ok {
		return x.Raw
	}
	return nil
}

func (x *Blob) GetZlibData() []byte {
	if x, ok := x.GetData().(*Blob_ZlibData); ok {
		return x.ZlibData
	}
	return nil
}

func (x *Blob) GetLzmaData() []byte {
	if x, ok := x.GetData().(*Blob_LzmaData); ok {
		return x.LzmaData
	}
	return nil
}

// Deprecated: Do not use.
func (x *Blob) GetOBSOLETEBzip2Data() []byte {
	if x, ok := x.GetData().(*Blob_OBSOLETEBzip2Data); ok {
		return x.OBSOLETEBzip2Data
	}
	return nil
}

func (x *Blob) GetLz4Data() []byte {
	if x, ok := x.GetData().(*Blob_Lz4Data); ok {
		return x.Lz4Data
	}
	return nil
}

func (x *Blob) GetZstdData() []byte {
	if x, ok := x.GetData().(*Blob_ZstdData); ok {
		return x.ZstdData
	}
	return nil
}

type isBlob_Data interface {
	isBlob_Data()
}

type Blob_Raw struct {
	Raw []byte `protobuf:"bytes,1,opt,name=raw,oneof"` // No compression
}

type Blob_ZlibData struct {
	// Possible compressed versions of the data.
	ZlibData []byte `protobuf:"bytes,3,opt,name=zlib_data,json=zlibData,oneof"`
}

type Blob_LzmaData struct {
	// For LZMA compressed data (optional)
	LzmaData []byte `protobuf:"bytes,4,opt,name=lzma_data,json=lzmaData,oneof"`
}

type Blob_OBSOLETEBzip2Data struct {
	// Formerly used for bzip2 compressed data. Deprecated in 2010.
	//
	// Deprecated: Do not use.
	OBSOLETEBzip2Data []byte `protobuf:"bytes,5,opt,name=OBSOLETE_bzip2_data,json=OBSOLETEBzip2Data,oneof"` // Don't reuse this tag number.
}

type Blob_Lz4Data struct {
	// For LZ4 compressed data (optional)
	Lz4Data []byte `protobuf:"bytes,6,opt,name=lz4_data,json=lz4Data,oneof"`
}

type Blob_ZstdData struct {
	// For ZSTD compressed data (optional)
	ZstdData []byte `protobuf:"bytes,7,opt,name=zstd_data,json=zstdData,oneof"`
}

func (*Blob_Raw) isBlob_Data() {}

func (*Blob_ZlibData) isBlob_Data() {}

func (*Blob_LzmaData) isBlob_Data() {}

func (*Blob_OBSOLETEBzip2Data) isBlob_Data() {}

func (*Blob_Lz4Data) isBlob_Data() {}

func (*Blob_ZstdData) isBlob_Data() {}

type BlobHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type      *string `protobuf:"bytes,1,req,name=type" json:"type,omitempty"`
	Indexdata []byte  `protobuf:"bytes,2,opt,name=indexdata" json:"indexdata,omitempty"`
	Datasize  *int32  `protobuf:"varint,3,req,name=datasize" json:"datasize,omitempty"`
}

func (x *BlobHeader) Reset() {
	*x = BlobHeader{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fileformat_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlobHeader) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlobHeader) ProtoMessage() {}

func (x *BlobHeader) ProtoReflect() protoreflect.Message {
	mi := &file_fileformat_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlobHeader.ProtoReflect.Descriptor instead.
func (*BlobHeader) Descriptor() ([]byte, []int) {
	return file_fileformat_proto_rawDescGZIP(), []int{1}
}

func (x *BlobHeader) GetType() string {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return ""
}

func (x *BlobHeader) GetIndexdata() []byte {
	if x != nil {
		return x.Indexdata
	}
	return nil
}

func (x *BlobHeader) GetDatasize() int32 {
	if x != nil && x.Datasize != nil {
		return *x.Datasize
	}
	return 0
}

var File_fileformat_proto protoreflect.FileDescriptor

var file_fileformat_proto_rawDesc = []byte{
	0x0a, 0x10, 0x66, 0x69, 0x6c, 0x65, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xed, 0x01, 0x0a, 0x04, 0x42, 0x6c, 0x6f, 0x62, 0x12, 0x19, 0x0a, 0x08, 0x72,
	0x61, 0x77, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x72,
	0x61, 0x77, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x12, 0x0a, 0x03, 0x72, 0x61, 0x77, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x03, 0x72, 0x61, 0x77, 0x12, 0x1d, 0x0a, 0x09, 0x7a, 0x6c,
	0x69, 0x62, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52,
	0x08, 0x7a, 0x6c, 0x69, 0x62, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1d, 0x0a, 0x09, 0x6c, 0x7a, 0x6d,
	0x61, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x08,
	0x6c, 0x7a, 0x6d, 0x61, 0x44, 0x61, 0x74, 0x61, 0x12, 0x34, 0x0a, 0x13, 0x4f, 0x42, 0x53, 0x4f,
	0x4c, 0x45, 0x54, 0x45, 0x5f, 0x62, 0x7a, 0x69, 0x70, 0x32, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x02, 0x18, 0x01, 0x48, 0x00, 0x52, 0x11, 0x4f, 0x42, 0x53,
	0x4f, 0x4c, 0x45, 0x54, 0x45, 0x42, 0x7a, 0x69, 0x70, 0x32, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1b,
	0x0a, 0x08, 0x6c, 0x7a, 0x34, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c,
	0x48, 0x00, 0x52, 0x07, 0x6c, 0x7a, 0x34, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1d, 0x0a, 0x09, 0x7a,
	0x73, 0x74, 0x64, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00,
	0x52, 0x08, 0x7a, 0x73, 0x74, 0x64, 0x44, 0x61, 0x74, 0x61, 0x42, 0x06, 0x0a, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x22, 0x5a, 0x0a, 0x0a, 0x42, 0x6c, 0x6f, 0x62, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72,
	0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x02, 0x28, 0x09, 0x52, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x64, 0x61,
	0x74, 0x61, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x61, 0x74, 0x61, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03,
	0x20, 0x02, 0x28, 0x05, 0x52, 0x08, 0x64, 0x61, 0x74, 0x61, 0x73, 0x69, 0x7a, 0x65,
}

var (
	file_fileformat_proto_rawDescOnce sync.Once
	file_fileformat_proto_rawDescData = file_fileformat_proto_rawDesc
)

func file_fileformat_proto_rawDescGZIP() []byte {
	file_fileformat_proto_rawDescOnce.Do(func() {
		file_fileformat_proto_rawDescData = protoimpl.X.CompressGZIP(file_fileformat_proto_rawDescData)
	})
	return file_fileformat_proto_rawDescData
}

var file_fileformat_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_fileformat_proto_goTypes = []interface{}{
	(*Blob)(nil),       // 0: Blob
	(*BlobHeader)(nil), // 1: BlobHeader
}
var file_fileformat_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_fileformat_proto_init() }
func file_fileformat_proto_init() {
	if File_fileformat_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_fileformat_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Blob); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_fileformat_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlobHeader); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_fileformat_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Blob_Raw)(nil),
		(*Blob_ZlibData)(nil),
		(*Blob_LzmaData)(nil),
		(*Blob_OBSOLETEBzip2Data)(nil),
		(*Blob_Lz4Data)(nil),
		(*Blob_ZstdData)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_fileformat_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_fileformat_proto_goTypes,
		DependencyIndexes: file_fileformat_proto_depIdxs,
		MessageInfos:      file_fileformat_proto_msgTypes,
	}.Build()
	File_fileformat_proto = out.File
	file_fileformat_proto_rawDesc = nil
	file_fileformat_proto_goTypes = nil
	file_fileformat_proto_depIdxs = nil
}
