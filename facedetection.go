package facedetection

import "errors"

var (
	ErrNoImageSpecified      = errors.New("no image specified")
	ErrBadInput              = errors.New("bad input data")
	ErrCantReadImage         = errors.New("can't read image")
	ErrImageTypeNotSupported = errors.New("image type not supported")
	ErrNoFacesFound          = errors.New("no faces found")
)

// FaceDetection represents detected faces information
type FaceDetection struct {
	Faces []Face `json:"faces"`
}

// Face represents face information with mouth and eyes coordinates
type Face struct {
	Bounds   FaceBounds `json:"bounds"`
	Mouth    Mouth      `json:"mouth"`
	LeftEye  Eye        `json:"left_eye"`
	RightEye Eye        `json:"right_eye"`
}

// FaceBounds represents face bounds information
type FaceBounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Height int `json:"height"`
	Width  int `json:"width"`
}

// Eye represents eye coordinates
type Eye struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Scale int `json:"scale"`
}

// Mouth represents mouth coordinates
type Mouth struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Height int `json:"height"`
	Width  int `json:"width"`
}
