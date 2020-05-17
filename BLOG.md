Write a short, yet technical blog post outlining how to build an image analysis API using Go and be sure to incorporate the following
Introduction
Directory and code structure
Common "gotchas"
Conclusion / Final Thoughts

Face Detection API in golang
=====

Face Detection service was developed with the idea to provide simple API to the end user which will allow detecting face, and it's parts like eyes and mouth.
This simple API can be used to build more advanced services like face blurring or face masks, etc.

Golang was chosen as a language to implement this service.

## Face detection library

First thing which comes to my mind when I'm thinking about image processing and computer vision is OpenCV. 
I was already working with this library when I was playing around with C++. Golang has package to use OpenCV 4, it can be found here: [gocv](https://github.com/hybridgroup/gocv).
Although, it requires additional software to be installed on your computer, so I've decided to take a look on other options, there are several of them:

* [go-face](https://github.com/Kagami/go-face)
* [pigo](https://github.com/esimov/pigo)
* [gocv](https://github.com/hybridgroup/gocv)

The only library which doesn't require any additional software is pigo. I like pure go libraries and decided to use this one as my main face detection library.

Now, as library is chosen we can implement the service.

## Project structure

I like those style guides of how to build golang packages:

* [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
* [Style guideline for Go packages](https://rakyll.org/style-packages/)

My project code structure looks next:

