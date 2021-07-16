package pipeline

import (
	"archive/tar"
	"io"
	"os"
	"strings"
)

func ReadFromTar(file string, heartbeat Heartbeat) (chan []byte, error) {
	open, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	channel := make(chan []byte, 10000)
	go func() {
		reader := tar.NewReader(open)
		heartbeat.Start()
		for {
			header, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				close(channel)
				panic(err)
			}
			if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, "geojson") {
				content := make([]byte, header.Size)
				_, err := reader.Read(content)
				if err != nil && err != io.EOF {
					close(channel)
					panic(err)
				}
				channel <- content
				heartbeat.Beat(1)
			}
		}
		close(channel)
		err := open.Close()
		heartbeat.Done()
		if err != nil {
			panic(err)
		}
	}()
	return channel, nil
}
