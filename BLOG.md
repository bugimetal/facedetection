Write a short, yet technical blog post outlining how to build an image analysis API using Go and be sure to incorporate the following
Introduction
Directory and code structure
Common "gotchas"
Conclusion / Final Thoughts

Face Detection API in golang
=====

Face Detection service was developed with the idea to provide simple API to the end user which allows detecting face, and it's parts like eyes and mouth.
This simple API can be used to build more advanced services like face blurring or face masks, etc.

Golang was chosen as a language to implement this service.

## Face detection library

First thing which comes to my mind when I'm thinking about image processing and computer vision is OpenCV. 
I was already working with this library when I was playing around with C++. Golang has package to use OpenCV 4, it can be found here: [gocv](https://github.com/hybridgroup/gocv).
Although, it requires additional software to be installed on your computer, so I've decided to take a look on other options, there are several of them:

* [go-face](https://github.com/Kagami/go-face)
* [pigo](https://github.com/esimov/pigo)
* [gocv](https://github.com/hybridgroup/gocv)

The only library which doesn't require any additional software is `pigo`. I like pure go libraries and decided to use this one as my main face detection library.

Now, as library is chosen we can implement the service.

## Project structure

When building the golang project I like to use best practices from those style guides:

* [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
* [Style guideline for Go packages](https://rakyll.org/style-packages/)

This project code structure looks following:

![project layout](web/blog/project_layout.png)

* `cmd/facedetection` - contains main application. 
The flow is simple, it creates the services, then it creates a handler and passes services as a dependency to it. After the handler is created, it listens for a new connections.
* `handler` - contains everything that is related to serving HTTP requests, like: router, error handling and http handlers itself.
Handler has services as a dependency, it uses them when processing the request. 
* `service` - contains services with main business logic. For this application we have two services: `ImageFetcher` and `FaceDetection` services. 
First is responsible for fetching the image from the internet by url and validating if the actual content is an image. Second service is responsible for actual face recognition. 
* `web` - contains static files for demo purposes.
* `facedetection.go` - in this file represented main structures that can be used by any parts of the application. Also, common errors that can be used by services specified here.
* `cascade` - this directory contains classifiers that needed for face detection library.

