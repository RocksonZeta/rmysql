package rmysql

type RMysqlError struct {
	Err error
	Msg string
}

func CheckError(err error) {
	if err != nil {
		panic(&RMysqlError{Err: err, Msg: ""})
	}
}
func Panic(msg string, err error) {
	panic(&RMysqlError{Err: err, Msg: msg})
}

func (r RMysqlError) Error() string {
	if r.Err != nil {
		return r.Msg + r.Err.Error()
	}
	return r.Msg
}
