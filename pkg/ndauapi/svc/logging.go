package svc

import (
	"net/http"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app"
	"github.com/sirupsen/logrus"
)

// LogWriter proxies http.ResponseWriter and logs.
type LogWriter struct {
	http.ResponseWriter
	status int
	length int
}

// WriteHeader proxies http.ResponseWriter.WriteHeader
func (w *LogWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// WriteHeader proxies http.ResponseWriter.Write
func (w *LogWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

// LogMW wraps a regular handler and replaces the writer with some logging middleware.
func LogMW(handler http.Handler) http.HandlerFunc {
	// Default logger that consults LOG_FORMAT and LOG_LEVEL env vars and logs to Stderr.
	logger := app.NewLogger()

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := LogWriter{ResponseWriter: w}
		handler.ServeHTTP(&lw, r)
		duration := time.Now().Sub(start)
		logger.WithFields(logrus.Fields{
			"host":       r.Host,
			"remoteAddr": r.RemoteAddr,
			"method":     r.Method,
			"uri":        r.RequestURI,
			"code":       lw.status,
			"len":        lw.length,
			"ua":         r.Header.Get("User-Agent"),
			"took":       duration,
		}).Info("REQ")
	}
}
