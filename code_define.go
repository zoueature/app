package app

type Errcode interface {
	Code() int
	Error() string
}

type ErrCode int

const (
	ApiStatusOK   ErrCode = 0
	ErrBadRequest ErrCode = 400
	ErrForbidden  ErrCode = 403
	ErrNotFound   ErrCode = 404
	ErrServer     ErrCode = 500
)

func (e ErrCode) Code() int {
	return int(e)
}

func (e ErrCode) Error() string {
	err := msgMap[e]
	return err
}

// RegisterCode 注入code message
func RegisterCode(codes map[ErrCode]string) {
	for code, msg := range codes {
		msgMap[code] = msg
	}

}

var msgMap = map[ErrCode]string{
	ErrBadRequest: "Bad Request",
	ErrForbidden:  "Access Denied",
	ErrNotFound:   "Not Found",
	ErrServer:     "Server Error",
	ApiStatusOK:   "OK",
}
