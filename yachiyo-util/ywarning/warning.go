package ywarning

type Warning interface {
	error
	Warning()
}
