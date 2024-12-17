package httperr

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"unsafe"
)

// EnsureSupported is a debug helper. Prints warning message is !ctrl.Supported(), with optional logger and stacktrace
// Prefer using EnsureMiddleware when possible.
//
//	 ctrl := httperr.EnsureSupported(httperr.NewResponseController(rw),
//	 	httperr.EnsureWithLogger(logger),
//		httperr.EnsureWithCallStack())
func EnsureSupported(ctrl *ResponseController, ops ...EnsureOp) *ResponseController {
	if !ctrl.Supported() {
		options := newEnsureOp(ops)
		options.warn(ctrl.rw)
	}
	return ctrl
}

func EnsureWithLogger(logger *slog.Logger) EnsureOp {
	return func(op *ensureOp) {
		op.logger = logger
	}
}

func EnsureWithCallStack() EnsureOp {
	return func(op *ensureOp) {
		op.withStack = true
	}
}

type EnsureOp func(op *ensureOp)

type ensureOp struct {
	logger    *slog.Logger
	withStack bool
}

func newEnsureOp(ops []EnsureOp) *ensureOp {
	options := &ensureOp{}
	for _, op := range ops {
		op(options)
	}
	if options.logger == nil {
		options.logger = slog.Default()
	}
	options.logger = options.logger.WithGroup("httperr")
	return options
}

func (o *ensureOp) warn(w http.ResponseWriter) {
	type_ := fmt.Sprintf("%T", w)
	msg := fmt.Sprintf("httperr: http.ResponseWriter[%s] neither implements httperr.ResponseWriter, nor wraps it.", type_)
	args := make([]any, 0, 4)
	args = append(args, "response_writer_type", type_)
	if o.withStack {
		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		stack := unsafe.String(unsafe.SliceData(buf), len(buf))
		args = append(args, "stack", stack)
	}
	o.logger.Warn(msg, args...)
}
