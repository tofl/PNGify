package main

import (
	"github.com/tofl/test/image"
)

func main() {
	img := image.NewImage([]byte("Hello world"))

	img.MakeImage()
}
