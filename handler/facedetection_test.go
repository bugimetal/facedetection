package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bugimetal/facedetection"
	"github.com/bugimetal/facedetection/service"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

func TestHandler_faceDetection(t *testing.T) {
	services, err := service.New()
	if err != nil {
		t.Fatalf("can't initialize services %v", err)
		return
	}

	h := New(Services{
		FaceDetection: services.FaceDetection,
		ImageFetcher:  services.ImageFetcher,
	})

	t.Run("no image provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := newRequest(http.MethodGet, "/v1/facedetection/bad_image_url", nil)

		h.faceDetection(w, r, []httprouter.Param{{Key: "image_url", Value: "bad_image_url"}})

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Expected HTTP status code %d, but got %d", http.StatusBadRequest, w.Code)
		}
	})

	astronautImageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("../web/test_image1.jpg")
		if err != nil {
			t.Fatalf("can't read image for test %v", err)
		}
		defer file.Close()
		img, _, _ := image.Decode(file)
		buf := new(bytes.Buffer)
		jpeg.Encode(buf, img, &jpeg.Options{100})

		w.Header().Set("Content-Type", "image/jpeg")
		if _, err := w.Write(buf.Bytes()); err != nil {
			logrus.Println("unable to write image.")
		}
	}))
	defer astronautImageServer.Close()

	t.Run("detecting faces on astronaut image", func(t *testing.T) {
		expectedFacedToBeDetected := 17
		imageUrlBase64 := base64.StdEncoding.EncodeToString([]byte(astronautImageServer.URL))

		w := httptest.NewRecorder()
		r := newRequest(http.MethodGet, fmt.Sprintf("/v1/facedetection/%s", imageUrlBase64), nil)

		h.faceDetection(w, r, []httprouter.Param{{Key: "image_url", Value: imageUrlBase64}})

		if w.Code != http.StatusOK {
			t.Fatalf("Expected HTTP status code %d, but got %d", http.StatusOK, w.Code)
		}

		var faces facedetection.FaceDetection
		if err := json.NewDecoder(w.Body).Decode(&faces); err != nil {
			t.Fatalf("can't decode response: %v", err)
		}

		if len(faces.Faces) != expectedFacedToBeDetected {
			t.Fatalf("Expected %d faces to be detected, but got %d", expectedFacedToBeDetected, len(faces.Faces))
		}
	})
}
