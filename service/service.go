package service

// Services contains all the services that this package has to offer.
type Services struct {
	*FaceDetection
	*ImageFetcher
}

// New returns Services.
func New() (*Services, error) {
	faceDetectionService, err := NewFaceDetection()
	if err != nil {
		return nil, err
	}

	return &Services{
		FaceDetection: faceDetectionService,
		ImageFetcher:  NewImageFetcher(),
	}, nil
}
