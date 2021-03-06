package service

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/bugimetal/facedetection"

	pigo "github.com/esimov/pigo/core"
	"github.com/sirupsen/logrus"
)

// FaceDetection service responsible for face detection
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

	// to read cascade files we need to build proper path
	_, b, _, _ := runtime.Caller(0)
	thisPackagePath := filepath.Dir(b)

	fd.faceCascade, err = ioutil.ReadFile(filepath.Join(thisPackagePath, "..", "cascade", "facefinder"))
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

	fd.puplocCascade, err = ioutil.ReadFile(filepath.Join(thisPackagePath, "..", "cascade", "puploc"))
	if err != nil {
		logrus.Errorf("Error reading the puploc cascade file: %s", err)
		return nil, err
	}
	fd.puplocClassifier, err = fd.puplocClassifier.UnpackCascade(fd.puplocCascade)
	if err != nil {
		logrus.Errorf("Error unpacking the puploc cascade file: %s", err)
		return nil, err
	}

	fd.flpCascades, err = fd.puplocClassifier.ReadCascadeDir(filepath.Join(thisPackagePath, "..", "cascade", "lps"))
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
			face.RightEye = facedetection.Eye{X: rightEye.Col, Y: rightEye.Row, Scale: int(rightEye.Scale / 8)}
		}

		leftEye := s.detectLeftPupil(faceCoords, imageParams)
		if leftEye != nil {
			face.LeftEye = facedetection.Eye{X: leftEye.Col, Y: leftEye.Row, Scale: int(leftEye.Scale / 8)}
		}

		if rightEye != nil && leftEye != nil {
			mouthPoints := s.detectMouthPoints(leftEye, rightEye, imageParams)
			p1, p2 := mouthPoints[0], mouthPoints[1]

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
		logrus.Warnf("cannot open the image: %v", err)
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
	flp1 := s.flpCascades["lp84"][0].FindLandmarkPoints(leftEye, rightEye, *imageParams, 63, false)
	flp2 := s.flpCascades["lp84"][0].FindLandmarkPoints(leftEye, rightEye, *imageParams, 63, true)
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
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
		ImageParams: *imageParams,
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	detections := s.faceClassifier.RunCascade(cascadeParams, 0)

	// Calculate the intersection over union (IoU) of two clusters.
	detections = s.faceClassifier.ClusterDetections(detections, 0.2)

	result := make([][]int, len(detections))
	var facesPassedThreshold int

	for i := 0; i < len(detections); i++ {
		// checking if model is sure enough about detected face
		if detections[i].Q > 5.0 {
			result[facesPassedThreshold] = append(result[i], detections[i].Row, detections[i].Col, detections[i].Scale, int(detections[i].Q))
			facesPassedThreshold++
		}
	}
	return result[:facesPassedThreshold]
}
