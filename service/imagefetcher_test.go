package service

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestImageFetcher_FetchImageByURL(t *testing.T) {
	type fields struct {
		client http.Client
	}
	type args struct {
		url string
	}

	jpegImageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := image.NewRGBA(image.Rect(0, 0, 240, 240))
		blue := color.RGBA{0, 0, 255, 255}
		draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

		var img image.Image = m
		buffer := new(bytes.Buffer)
		if err := jpeg.Encode(buffer, img, nil); err != nil {
			log.Println("unable to encode image.")
		}

		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
		if _, err := w.Write(buffer.Bytes()); err != nil {
			logrus.Println("unable to write image.")
		}
	}))
	defer jpegImageServer.Close()

	helloWorldServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world!")
	}))
	defer helloWorldServer.Close()

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    io.Reader
		wantErr bool
	}{
		{
			name:    "jpeg image url provided",
			args:    args{url: jpegImageServer.URL},
			wantErr: false,
		},
		{
			name:    "no image provided",
			args:    args{url: helloWorldServer.URL},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ImageFetcher{
				client: tt.fields.client,
			}
			_, err := s.FetchImageByURL(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchImageByURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
