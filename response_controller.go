package httperr

import "net/http"

func NewResponseController(rw http.ResponseWriter) *ResponseController {
	impl := unwrap(rw)
	return &ResponseController{impl}
}

type ResponseController struct {
	rw ResponseWriter
}

// Error adds err to the list. If method is not supported, returns (nil, false)
func (r *ResponseController) Error(err error) (error, bool) {
	if r.rw == nil {
		return nil, false
	}
	return r.rw.Error(err), true
}

// Errors returns list of collected errors. If method is not supported, returns (nil, false)
func (r *ResponseController) Errors() ([]error, bool) {
	if r.rw == nil {
		return nil, false
	}
	return r.rw.Errors(), true
}

// Supported returns if [ResponseWriter] is supported by wrapped [http.ResponseWriter]
func (r *ResponseController) Supported() bool {
	return r.rw != nil
}

type rwUnwrapper interface {
	Unwrap() http.ResponseWriter
}

func unwrap(rw http.ResponseWriter) ResponseWriter {
	for {
		switch t := rw.(type) {
		case ResponseWriter:
			return t
		case rwUnwrapper:

			rw = t.Unwrap()
		default:
			return nil
		}
	}
}
