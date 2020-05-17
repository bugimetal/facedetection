package service

import (
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"path/filepath"

	pigo "github.com/esimov/pigo/core"
	"github.com/sirupsen/logrus"

	"github.com/bugimetal/facedetection"
)

type FaceDetection struct {
	faceCascade   []byte
	puplocCascade []byte

	faceClassifier   *pigo.Pigo
	puplocClassifier *pigo.PuplocCascade

	flpCascades map[string][]*pigo.FlpCascade
}

// NewFaceDetection returns new Face Detection service with prebuild cascades
func NewFaceDetection() (*FaceDetection, error) {
	var err error
	fd := &FaceDetection{}

	fd.faceCascade, err = ioutil.ReadFile(filepath.Join("cascade", "facefinder"))
	if err != nil {
		logrus.Errorf("error reading the cascade file: %v", err)
		return nil, err
	}
	p := pigo.NewPigo()

	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	fd.faceClassifier, err = p.Unpack(fd.faceCascade)
	if err != nil {
		logrus.Errorf("Error unpacking the cascade file: %s", err)
		return nil, err
	}

	fd.puplocCascade, err = ioutil.ReadFile(filepath.Join("cascade", "puploc"))
	if err != nil {
		logrus.Errorf("Error reading the puploc cascade file: %s", err)
		return nil, err
	}
	fd.puplocClassifier, err = fd.puplocClassifier.UnpackCascade(fd.puplocCascade)
	if err != nil {
		logrus.Errorf("Error unpacking the puploc cascade file: %s", err)
		return nil, err
	}

	fd.flpCascades, err = fd.puplocClassifier.ReadCascadeDir(filepath.Join("cascade", "lps"))
	if err != nil {
		logrus.Errorf("Error unpacking the facial landmark detection cascades: %s", err)
		return nil, err
	}

	return fd, nil
}

// Detect detects faces and returns face coordinates along with eyes and mouth
func (s *FaceDetection) Detect(image io.Reader) (facedetection.FaceDetection, error) {
	var detectedFaces facedetection.FaceDetection

	imageParams, err := s.buildImageParams(image)
	if err != nil {
		return detectedFaces, err
	}

	foundFaces := s.findFaces(imageParams)
	if len(foundFaces) == 0 {
		return detectedFaces, facedetection.ErrNoFacesFound
	}

	// TODO: is this needed?
	detectedFaces.Faces = make([]facedetection.Face, 0)

	for _, faceCoords := range foundFaces {
		row, col, scale := faceCoords[1], faceCoords[0], faceCoords[2]
		face := facedetection.Face{
			Bounds: facedetection.FaceBounds{
				X:      row - scale/2,
				Y:      col - scale/2,
				Height: scale,
				Width:  scale,
			},
		}

		rightEye := s.detectRightPupil(faceCoords, imageParams)
		if rightEye != nil {
			col, row, scale = rightEye.Col, rightEye.Row, int(rightEye.Scale/8)
			face.RightEye = facedetection.Eye{X: col, Y: row, Scale: scale}
		}

		leftEye := s.detectLeftPupil(faceCoords, imageParams)
		if leftEye != nil {
			col, row, scale = leftEye.Col, leftEye.Row, int(leftEye.Scale/8)
			face.LeftEye = facedetection.Eye{X: col, Y: row, Scale: scale}
		}

		if rightEye != nil && leftEye != nil {
			mouth := s.detectMouthPoints(leftEye, rightEye, imageParams)
			p1, p2 := mouth[0], mouth[1]

			width := p2[0] - p1[0]
			height := p2[1] - p1[1]
			if height <= 0 {
				height = 1
			}

			face.Mouth = facedetection.Mouth{
				X:      p1[0],
				Y:      p1[1] + (p1[1]-p2[1])/2,
				Height: height,
				Width:  width,
			}

			detectedFaces.Faces = append(detectedFaces.Faces, face)
		}
	}

	if len(detectedFaces.Faces) == 0 {
		return detectedFaces, facedetection.ErrNoFacesFound
	}

	return detectedFaces, nil
}

// buildImageParams builds image parameters for pigo library
func (s *FaceDetection) buildImageParams(image io.Reader) (*pigo.ImageParams, error) {
	src, err := pigo.DecodeImage(image)
	if err != nil {
		logrus.Warnf("Cannot open the image file: %v", err)
		return nil, facedetection.ErrCantReadImage
	}

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	return &pigo.ImageParams{
		Pixels: pixels,
		Rows:   rows,
		Cols:   cols,
		Dim:    cols,
	}, nil
}

// detectLeftPupil detects the left pupil based on previously detected face coordinates
func (s *FaceDetection) detectLeftPupil(face []int, imageParams *pigo.ImageParams) *pigo.Puploc {
	puploc := &pigo.Puploc{
		Row:      face[0] - int(0.085*float32(face[2])),
		Col:      face[1] - int(0.185*float32(face[2])),
		Scale:    float32(face[2]) * 0.45,
		Perturbs: 50,
	}
	leftEye := s.puplocClassifier.RunDetector(*puploc, *imageParams, 0.0, false)
	if leftEye.Row > 0 && leftEye.Col > 0 {
		return leftEye
	}
	return nil
}

// detectRightPupil detects the right pupil based on previously detected face coordinates
func (s *FaceDetection) detectRightPupil(face []int, imageParams *pigo.ImageParams) *pigo.Puploc {
	puploc := &pigo.Puploc{
		Row:      face[0] - int(0.085*float32(face[2])),
		Col:      face[1] + int(0.185*float32(face[2])),
		Scale:    float32(face[2]) * 0.45,
		Perturbs: 50,
	}
	rightEye := s.puplocClassifier.RunDetector(*puploc, *imageParams, 0.0, false)
	if rightEye.Row > 0 && rightEye.Col > 0 {
		return rightEye
	}
	return nil
}

// detectMouthPoints detect mouth points
func (s *FaceDetection) detectMouthPoints(leftEye, rightEye *pigo.Puploc, imageParams *pigo.ImageParams) [][]int {
	flp1 := s.flpCascades["lp84"][0].FindLandmarkPoints(leftEye, rightEye, *imageParams, 50, false)
	flp2 := s.flpCascades["lp84"][0].FindLandmarkPoints(leftEye, rightEye, *imageParams, 50, true)
	return [][]int{
		[]int{flp1.Col, flp1.Row, int(flp1.Scale)},
		[]int{flp2.Col, flp2.Row, int(flp2.Scale)},
	}
}

// findFaces finds all faces coordinates
func (s *FaceDetection) findFaces(imageParams *pigo.ImageParams) [][]int {
	cascadeParams := pigo.CascadeParams{
		MinSize:     50,
		MaxSize:     imageParams.Cols,
		ShiftFactor: 0.5,
		ScaleFactor: 1.02,
		ImageParams: *imageParams,
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	detections := s.faceClassifier.RunCascade(cascadeParams, 0)

	// Calculate the intersection over union (IoU) of two clusters.
	detections = s.faceClassifier.ClusterDetections(detections, 0.2)

	result := make([][]int, len(detections))

	for i := 0; i < len(detections); i++ {
		result[i] = append(result[i], detections[i].Row, detections[i].Col, detections[i].Scale, int(detections[i].Q))
	}
	return result
}
