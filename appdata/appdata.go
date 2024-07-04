package appdata

import (
	"webvid/vision"

	"gocv.io/x/gocv"
)

const (
	V4LCAM = iota
	ESPCAM
)

type AppData struct {
	Cameras []*vision.Camera
}

func NewAppData() *AppData {
	var data = &AppData{
		Cameras: []*vision.Camera{
			vision.NewCamera(2, gocv.VideoCaptureV4L),
			vision.NewCamera("http://192.168.0.28:8080", gocv.VideoCaptureAny),
		},
	}
	return data
}

func (data *AppData) Load() {
}
