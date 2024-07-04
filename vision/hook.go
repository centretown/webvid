package vision

import "gocv.io/x/gocv"

type Hook interface {
	Update(img *gocv.Mat)
	Close(int)
}

type UiHook interface {
	Hook
	SetUi(ui interface{})
}
