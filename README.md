Face Detections
======
[![Go Report Card](https://goreportcard.com/badge/github.com/bugimetal/facedetection)](https://goreportcard.com/report/github.com/bugimetal/facedetection)

Face Detection service provides simple API to detect faces on given image. 
Service will also detect eyes (pupil) points and mouth area for every detected face. 

## 1. The API.

The actual API has only 1 endpoint which requires no authentication.

|Method|Endpoint                                     |Description                                           |
|------|---------------------------------------------|------------------------------------------------------|
|GET   |/v1/facedetection/:image_url_base64_encoded* |Detect faces on given image and returns json response |

*provided image URL should be base64 encoded

Example request:
```
curl http://localhost:8080/v1/facedetection/aHR0cHM6Ly9yYXcuZ2l0aHVidXNlcmNvbnRlbnQuY29tL2VzaW1vdi9waWdvL21hc3Rlci90ZXN0ZGF0YS9zYW1wbGUuanBn
```

Example response:
```
{
    "faces": [
        {
            "bounds": {
                "x": 573,
                "y": 79,
                "height": 52,
                "width": 52
            },
            "mouth": {
                "x": 596,
                "y": 125,
                "height": 1,
                "width": 18
            },
            "left_eye": {
                "x": 595,
                "y": 103,
                "scale": 0
            },
            "right_eye": {
                "x": 615,
                "y": 103,
                "scale": 0
            }
        },
        ...
    ]
}
```

## 2. How to run service locally.

It's super simple, service has no configuration required to start.

```
go run ./cmd/facedetection/
```

## 3. Demo.

In order to see demo for the service, [run the service locally](#2-how-to-run-service-locally) and visit [Demo page](http://localhost:8080/web/demo.html)

## 4. How to run tests.

`go test -v github.com/bugimetal/facedetection/...`