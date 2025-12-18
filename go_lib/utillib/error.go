package utillib

import "fmt"

type Err struct {
	ErrCode int32
	ErrMsg  string
}

func (e Err) Error() string {
	return fmt.Sprintf("ErrCode: %v ErrMsg: %v", e.ErrCode, e.ErrMsg)
}

func NewErr(errCode int32, errMsg string) error {
	return &Err{
		ErrCode: errCode,
		ErrMsg:  errMsg,
	}
}

func GetErrorInfo(err error) (int32, string) {
	switch e := err.(type) {
	case *Err:
		return e.ErrCode, e.ErrMsg
	}
	return -1, err.Error()
}
