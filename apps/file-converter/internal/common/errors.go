package common

type SimpleErr struct {
	Msg string
}

type NotFoundErr struct {
	SimpleErr
}

type NotAllowedErr struct {
	SimpleErr
}

type IllegalStateErr struct {
	SimpleErr
}

func (e SimpleErr) Error() string {
	return e.Msg
}
