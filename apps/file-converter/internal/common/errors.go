package common

type NotFoundErr struct {
	Msg string
}

type IllegalStateErr struct {
	Msg string
}

func (e NotFoundErr) Error() string {
	return e.Msg
}

func (e IllegalStateErr) Error() string {
	return e.Msg
}
