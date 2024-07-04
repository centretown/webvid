package vision

import (
	"fmt"
	"log"
	"time"

	"gocv.io/x/gocv"
)

type Verb uint16

const (
	GET Verb = iota
	SET
	HIDEALL
	RECORD_START
	RECORD_STOP
)

const (
	RecordingFolder = "recordings/"
)

var cmdList = []string{
	"Get",
	"Set",
	"HideAll",
}

func (cmd Verb) String() string {
	if cmd >= Verb(len(cmdList)) {
		return "Unknown"
	}
	return cmdList[cmd]
}

type CameraCmd struct {
	Action   Verb
	Property gocv.VideoCaptureProperties
	Value    any
}

type Camera struct {
	ID   any
	API  gocv.VideoCaptureAPI
	Name string

	Quit chan int
	Cmd  chan CameraCmd

	recordStop time.Time

	StreamHook *StreamHook

	Filters []Hook

	HideMain  bool
	HideThumb bool
	HideAll   bool
	Busy      bool
	Recording bool

	FrameWidth  float64
	FrameHeight float64
	FrameRate   float64

	video  *gocv.VideoCapture
	writer *gocv.VideoWriter

	matrix gocv.Mat
}

func NewCamera(id any, api gocv.VideoCaptureAPI) *Camera {
	name, ok := id.(string)
	if !ok {
		num, _ := id.(int)
		name = fmt.Sprintf("V4L-%02d", num)
	}

	cam := &Camera{
		ID:         id,
		API:        api,
		Name:       name,
		Quit:       make(chan int),
		Cmd:        make(chan CameraCmd),
		StreamHook: NewStreamHook(),
		Filters:    make([]Hook, 0),

		FrameWidth:  1280,
		FrameHeight: 720,
		FrameRate:   20,
		writer:      nil,
	}

	return cam
}

func (cam *Camera) AddFilter(filter Hook) {
	cam.Filters = append(cam.Filters, filter)
}
func (cam *Camera) Command(cmd CameraCmd) {
	cam.Cmd <- cmd
}

func (cam *Camera) RecordCmd() {
	cam.Command(CameraCmd{Action: RECORD_START, Value: true})
}

func (cam *Camera) StopRecordCmd() {
	cam.Command(CameraCmd{Action: RECORD_STOP, Value: true})
}

func (cam *Camera) Open() (err error) {
	var (
		useAPI = cam.API > 0
	)
	if useAPI {
		cam.video, err = gocv.OpenVideoCaptureWithAPI(cam.ID, cam.API)
	} else {
		cam.video, err = gocv.OpenVideoCapture(cam.ID)
	}

	if err != nil {
		log.Println(err, cam.ID, "OpenVideoCapture")
		return
	}

	if useAPI {
		cam.video.Set(gocv.VideoCaptureFPS, cam.FrameRate)
		cam.video.Set(gocv.VideoCaptureFrameHeight, cam.FrameHeight)
		cam.video.Set(gocv.VideoCaptureFrameWidth, cam.FrameWidth)
	}

	cam.FrameWidth = cam.video.Get(gocv.VideoCaptureFrameWidth)
	cam.FrameHeight = cam.video.Get(gocv.VideoCaptureFrameHeight)
	cam.FrameRate = cam.video.Get(gocv.VideoCaptureFPS)
	log.Printf("Opened '%s' Size: %.0fx%.0f FPS: %.0f\n", cam.Name, cam.FrameWidth, cam.FrameHeight, cam.FrameRate)
	return
}

func (cam *Camera) Close() {
	if cam.writer != nil && cam.writer.IsOpened() {
		cam.writer.Close()
	}
	cam.StreamHook.Close(0)
	cam.matrix.Close()
	cam.video.Close()
	log.Printf("Closed '%s'\n", cam.Name)
}

const (
	delayNormal    = time.Millisecond * 20
	delayRetry     = time.Second
	delayHibernate = time.Second * 30
	recordLimit    = time.Second * 5
)

func (cam *Camera) stopRecording() {
	cam.Recording = false
	if cam.writer == nil {
		return
	}

	defer func() {
		cam.writer = nil
	}()

	if !cam.writer.IsOpened() {
		log.Println("writer already closed")
		return
	}

	log.Println("close writer")
	err := cam.writer.Close()
	if err != nil {
		log.Println("stop recording", err)
		return
	}
	log.Println("recorder closed")
}

func timeName() string {
	now := time.Now()
	return fmt.Sprintf("record_%d_%d_%d_%d_%d_%d.mp4", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
}

func (cam *Camera) startRecording(duration int) {
	var err error
	log.Println("start recording")

	if cam.Recording {
		log.Println("already recording")
		cam.stopRecording()
		return
	}

	// avc1	- H.264
	cam.writer, err = gocv.VideoWriterFile(timeName(), "avc1", 10, 1280, 720, true)
	if err != nil {
		log.Println("start recording failed", err)
		return
	}

	cam.Recording = true
	now := time.Now()
	cam.recordStop = now.Add(time.Second * time.Duration(duration))
	log.Println("recording started...")

}

func (cam *Camera) doCmd(cmd CameraCmd) {
	switch cmd.Action {
	case GET:
		cmd.Value = cam.video.Get(cmd.Property)
	case SET:
		f, _ := cmd.Value.(float64)
		cam.video.Set(cmd.Property, float64(f))
	case HIDEALL:
		b, _ := cmd.Value.(bool)
		cam.HideAll = b
	case RECORD_START:
		cam.startRecording(cmd.Value.(int))
	case RECORD_STOP:
		cam.stopRecording()
	}
}

func (cam *Camera) Serve() {
	if cam.Busy {
		return
	}

	var (
		cmd   CameraCmd
		retry int = 0
	)

	err := cam.Open()
	if err != nil {
		return
	}

	cam.Busy = true

	defer func() {
		cam.Busy = false
		cam.Close()
	}()

	var (
		delay = delayNormal
	)

	cam.matrix = gocv.NewMat()

	for {
		time.Sleep(delay)

		select {
		case <-cam.Quit:
			return
		case cmd = <-cam.Cmd:
			cam.doCmd(cmd)
			continue
		default:
		}

		if cam.HideAll {
			continue
		}

		if !cam.video.Read(&cam.matrix) {
			if retry > 10 {
				delay = delayHibernate
			} else {
				delay = delayRetry
			}

			log.Printf("%v is unavailable, attempts=%d next in %.0f seconds\n",
				cam.ID, retry, delay.Seconds())
			retry++
			if cam.video.IsOpened() {
				cam.video.Close()
			}
			cam.Open()
			continue
		}
		delay = delayNormal
		retry = 0

		if cam.matrix.Empty() {
			continue
		}

		for _, filter := range cam.Filters {
			filter.Update(&cam.matrix)
		}

		cam.StreamHook.Update(&cam.matrix)

		if cam.writer != nil && cam.Recording && cam.writer.IsOpened() {
			err := cam.writer.Write(cam.matrix)
			if err != nil {
				log.Println(err, "Write")
			}

			if cam.recordStop.Before(time.Now()) {
				cam.doCmd(CameraCmd{Action: RECORD_STOP})
			}
		}
	}

}
