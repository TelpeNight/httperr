package httperr

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter

	// Error adds new err to [ResponseWriter] list. Returns error that was added (like gin does)
	Error(err error) error

	// Errors returns list of collected errors.
	Errors() []error

	// Unwrap returns original [http.ResponseWriter]. This module and all implementations guarantee to be compatible with [http.ResponseController]
	Unwrap() http.ResponseWriter
}

// ResponseWriterWithErrors ensures that resulting [http.ResponseWriter] is supported by [ResponseController].
// It returns original [http.ResponseWriter] is it already wraps [ResponseWriter],
// or creates new one.
func ResponseWriterWithErrors(rw http.ResponseWriter) http.ResponseWriter {
	if supported := unwrap(rw); supported != nil {
		return rw
	}
	return wrap(rw)
}

func wrap(rw http.ResponseWriter) ResponseWriter {
	writer := responseWriter{
		ResponseWriter: rw,
		ctrl:           http.NewResponseController(rw),
	}
	if _, is := rw.(io.ReaderFrom); is {
		return &readerFromWriter{writer}
	}
	return &writer
}

type responseWriter struct {
	http.ResponseWriter
	ctrl *http.ResponseController
	errs []error
}

type readerFromWriter struct {
	responseWriter
}

var (
	_ ResponseWriter = (*responseWriter)(nil)
	// also implement some well-known interfaces to be compatible with pre-1.22 use-cases
	_ http.Hijacker   = (*responseWriter)(nil)
	_ http.Flusher    = (*responseWriter)(nil)
	_ io.StringWriter = (*responseWriter)(nil)
	_ io.ReaderFrom   = (*readerFromWriter)(nil)
)

func (w *responseWriter) Error(err error) error {
	w.errs = append(w.errs, err)
	return err
}

func (w *responseWriter) Errors() []error {
	return w.errs
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ctrl.Hijack()
}

func (w *responseWriter) Flush() {
	_ = w.ctrl.Flush()
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	return io.WriteString(w.ResponseWriter, s)
}

func (w *readerFromWriter) ReadFrom(r io.Reader) (n int64, err error) {
	return w.ResponseWriter.(io.ReaderFrom).ReadFrom(r)
}
