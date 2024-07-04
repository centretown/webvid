package main

import (
	"flag"
	"webvid/appdata"
	"webvid/web"
)

func main() {
	flag.Parse()
	data := appdata.NewAppData()
	web.Serve(data.Cameras)
}
