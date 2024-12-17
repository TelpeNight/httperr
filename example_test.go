package httperr_test

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TelpeNight/httperr"
)

func exampleTest(t *testing.T, handler http.Handler) {
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	discardResp(resp)
}

func TestLogErrors(t *testing.T) {
	handler := httperr.WithMiddleware(example500Handler(t),
		httperr.Middleware(),
		loggingMiddleware(t),
		httperr.EnsureMiddleware(httperr.EnsureWithCallStack()))
	exampleTest(t, handler)
}

func TestWrappedResponseWriter(t *testing.T) {
	handler := httperr.WithMiddleware(example500Handler(t),
		httperr.Middleware(),
		wrappingMiddleware,
		loggingMiddleware(t),
		httperr.EnsureMiddleware(httperr.EnsureWithCallStack()))
	exampleTest(t, handler)
}

func TestWrappedResponseWriterBetween(t *testing.T) {
	handler := httperr.WithMiddleware(example500Handler(t),
		httperr.Middleware(),
		loggingMiddleware(t),
		wrappingMiddleware,
		httperr.EnsureMiddleware(httperr.EnsureWithCallStack()))
	exampleTest(t, handler)
}

func TestStandaloneLogger(t *testing.T) {
	handler := httperr.WithMiddleware(example500Handler(t),
		standaloneLoggingMiddleware(t),
		httperr.EnsureMiddleware(httperr.EnsureWithCallStack()))
	exampleTest(t, handler)
}

func TestStandaloneLoggerWithDefault(t *testing.T) {
	handler := httperr.WithMiddleware(example500Handler(t),
		httperr.Middleware(),
		loggingMiddleware(t), // ensure the same error will be observed
		wrappingMiddleware,
		standaloneLoggingMiddleware(t),
		httperr.EnsureMiddleware(httperr.EnsureWithCallStack()))
	exampleTest(t, handler)
}

func TestEnsure(t *testing.T) {
	handler := httperr.WithMiddleware(example500Handler(t, true),
		httperr.EnsureMiddleware(httperr.EnsureWithCallStack()))
	exampleTest(t, handler)
}

func TestEnsureWithLegacyMiddleware(t *testing.T) {
	handler := httperr.WithMiddleware(example500Handler(t, true),
		httperr.Middleware(),
		loggingMiddleware(t),
		legacyWrappingMiddleware,
		httperr.EnsureMiddleware(httperr.EnsureWithCallStack()))
	exampleTest(t, handler)
}

func example500Handler(t *testing.T, willFail ...bool) http.Handler {
	expectFail := false
	if len(willFail) > 0 {
		expectFail = willFail[0]
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctrl := httperr.NewResponseController(w)
		if !ctrl.Supported() && !expectFail {
			t.Error("!ctrl.Supported()")
		} else if ctrl.Supported() && expectFail {
			t.Error("ctrl.Supported()")
		}
		_, _ = ctrl.Error(fmt.Errorf("example db connection error"))
		w.WriteHeader(http.StatusInternalServerError)
	})
}

func loggingMiddleware(t *testing.T) httperr.M {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			ctrl := httperr.NewResponseController(w)
			if !ctrl.Supported() {
				t.Error("!ctrl.Supported()")
			}
			errs, _ := ctrl.Errors()
			if len(errs) > 0 {
				err := errors.Join(errs...)
				slog.Error(err.Error())
			}
		})
	}
}

func standaloneLoggingMiddleware(t *testing.T) httperr.M {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w = httperr.ResponseWriterWithErrors(w)
			next.ServeHTTP(w, r)
			ctrl := httperr.NewResponseController(w)
			if !ctrl.Supported() {
				t.Error("!ctrl.Supported()")
			}
			errs, _ := ctrl.Errors()
			if len(errs) > 0 {
				err := errors.Join(errs...)
				slog.Error(err.Error())
			}
		})
	}
}

func wrappingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w = &rwWithUnwrap{w}
		next.ServeHTTP(w, r)
	})
}

func legacyWrappingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w = &rwOnly{w}
		next.ServeHTTP(w, r)
	})
}

type rwOnly struct {
	http.ResponseWriter
}

type rwWithUnwrap struct {
	http.ResponseWriter
}

func (rw *rwWithUnwrap) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func discardResp(resp *http.Response) {
	if resp == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}
