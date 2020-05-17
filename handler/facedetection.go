package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bugimetal/facedetection"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

func (handler *Handler) faceDetection(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	imageURLB64 := strings.TrimSpace(ps.ByName("image_url"))
	if imageURLB64 == "" {
		handler.Error(w, r, facedetection.ErrNoImageSpecified)
		return
	}

	imageURL, err := base64.StdEncoding.DecodeString(imageURLB64)
	if err != nil {
		handler.Error(w, r, facedetection.ErrBadInput)
		return
	}

	image, err := handler.imageFetcherService.FetchImageByURL(string(imageURL))
	if err != nil {
		handler.Error(w, r, err)
		return
	}

	detectedFaces, err := handler.faceDetectionService.Detect(image)
	if err != nil {
		handler.Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(detectedFaces); err != nil {
		logrus.Errorf("Unable to respond with detected faces %s", err)
	}
}
