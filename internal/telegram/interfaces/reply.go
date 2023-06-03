package interfaces

import "io"

type Replier interface {
	InternalError()
	Usage()
	ReplyWithMessage(msg string)
	SensitivePicture(pic io.Reader)
}
