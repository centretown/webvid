package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"webvid/vision"
)

var _ http.Handler = (*HandleRecord)(nil)

type HandleRecord struct {
	cam *vision.Camera
}

func NewHandleRecord(cam *vision.Camera) *HandleRecord {
	hr := &HandleRecord{
		cam: cam,
	}
	return hr
}

func (hr *HandleRecord) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if hr.cam == nil {
		log.Println("no camera")
		return
	}
	log.Println("record requested", r.Host)
	if !hr.cam.Busy {
		log.Println("cam is idle", r.Host)
		return
	}

	duration := 5
	values := r.URL.Query()
	parm, ok := values["duration"]
	if ok && len(parm) > 0 {
		i, err := strconv.Atoi(parm[0])
		if err == nil {
			duration = i
		}
	}

	log.Println("request values", len(values), values)
	for k, v := range values {
		log.Println("request values", k, v)
	}

	var (
		cmd     vision.CameraCmd
		message string
	)
	if hr.cam.Recording {
		cmd.Action = vision.RECORD_STOP
		message = "stop"
	} else {
		cmd.Action = vision.RECORD_START
		message = fmt.Sprintln("record for", duration, "seconds")
	}

	cmd.Value = duration
	hr.cam.Cmd <- cmd

	w.Write([]byte(message))
}
