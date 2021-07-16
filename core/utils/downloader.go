package utils

import (
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"os"
)

func DownloadFile(filepath string, url string) error {

	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	f, _ := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)
	_, err := io.Copy(io.MultiWriter(f, bar), resp.Body)
	return err
}
