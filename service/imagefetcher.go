package service

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bugimetal/facedetection"
)

var (
	supportedTypes = map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
	}
)

type ImageFetcher struct {
	client http.Client
}

func NewImageFetcher() *ImageFetcher {
	return &ImageFetcher{
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *ImageFetcher) FetchImageByURL(url string) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, facedetection.ErrCantReadImage
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, facedetection.ErrCantReadImage
	}

	t := http.DetectContentType(b)
	if _, ok := supportedTypes[t]; !ok {
		return nil, facedetection.ErrImageTypeNotSupported
	}

	return bytes.NewReader(b), nil
}
