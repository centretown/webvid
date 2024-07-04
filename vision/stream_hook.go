package vision

import (
	"log"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

type StreamHook struct {
	Stream *mjpeg.Stream
}

func NewStreamHook() *StreamHook {
	sh := &StreamHook{}
	sh.Stream = mjpeg.NewStream()
	return sh
}

var _ Hook = (*StreamHook)(nil)

func (sh *StreamHook) Update(img *gocv.Mat) {
	buf, err := gocv.IMEncode(".jpg", *img)
	if err != nil {
		log.Println("IMEncode", err)
		return
	}

	sh.Stream.UpdateJPEG(buf.GetBytes())
	buf.Close()
}

func (sh *StreamHook) Close(int) {}
