package vision

import (
	"image/color"
	"log"

	"gocv.io/x/gocv"
)

type ClassifyHook struct {
	classifier gocv.CascadeClassifier
}

var _ Hook = (*ClassifyHook)(nil)

func NewClassifyHook() *ClassifyHook {
	ch := &ClassifyHook{}
	ch.classifier = gocv.NewCascadeClassifier()
	if !ch.classifier.Load("classifier/haarcascade_frontalface_default.xml") {
		log.Println("Error reading cascade file: classifier/haarcascade_frontalface_default.xml")
	}
	return ch
}

var blue = color.RGBA{R: 0, G: 128, B: 255, A: 255}

func (ch *ClassifyHook) Update(img *gocv.Mat) {
	rects := ch.classifier.DetectMultiScale(*img)
	for _, r := range rects {
		gocv.Rectangle(img, r, blue, 2)
	}
}

func (ch *ClassifyHook) Close(int) {
	ch.classifier.Close()
}
