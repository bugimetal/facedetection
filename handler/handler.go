package handler

import (
	"io"
	"net/http"

	"github.com/bugimetal/facedetection"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

// FaceDetectionService provides an interface to the service that deals with face detection
type FaceDetectionService interface {
	Detect(io.Reader) (facedetection.FaceDetection, error)
}

// ImageFetcherService provides an interface to the service that deals with fetching images
type ImageFetcherService interface {
	FetchImageByURL(string) (io.Reader, error)
}

// Services describe the external services that the Handler relies on.
type Services struct {
	FaceDetection FaceDetectionService
	ImageFetcher  ImageFetcherService
}

// Handler provides an generic interface for handling HTTP requests.
type Handler struct {
	http                 http.Handler
	faceDetectionService FaceDetectionService
	imageFetcherService  ImageFetcherService
}

// New returns a new Handler.
func New(services Services) *Handler {
	handler := &Handler{
		faceDetectionService: services.FaceDetection,
		imageFetcherService:  services.ImageFetcher,
	}

	// Set up a custom HTTP router and install the routes on it.
	router := httprouter.New()

	router.GET("/v1/facedetection/:image_url", handler.faceDetection)

	// Serve static demo files
	router.ServeFiles("/static/*filepath", http.Dir("./static/"))

	handler.http = cors.Default().Handler(router)

	return handler
}

// ServeHTTP handles every incoming HTTP request and passes the request along
// to the configured HTTP router.
func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.http.ServeHTTP(w, r)
}
