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

// ImageFetcher service responsible for fetching images
type ImageFetcher struct {
	client http.Client
}

// NewImageFetcher new ImageFetcher service which is responsible for fetching and validation images
func NewImageFetcher() *ImageFetcher {
	return &ImageFetcher{
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// FetchImageByURL fetches image content by given URL. Validates if the downloaded file has suitable image format
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
