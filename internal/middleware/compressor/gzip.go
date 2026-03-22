package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

var allowedContentTypes = []string{
	"application/json",
	"text/html",
}

func shouldCompress(ct string) bool {
	if ct == "" {
		return false
	}

	for _, allowed := range allowedContentTypes {
		if strings.Contains(ct, allowed) {
			return true
		}
	}
	return false
}

type gzipWriter struct {
	http.ResponseWriter
	writer          io.Writer
	gz              *gzip.Writer
	clientSupportGz bool
}

func WithGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(r.Body)
			if err == nil {
				defer reader.Close()
				r.Body = reader
			}
		}

		clientSupportGz := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		gzw := &gzipWriter{
			ResponseWriter:  w,
			writer:          w,
			clientSupportGz: clientSupportGz,
		}

		next.ServeHTTP(gzw, r)

		if gzw.gz != nil {
			gzw.gz.Close()
		}
	})
}

func (w *gzipWriter) WriteHeader(statusCode int) {
	if w.clientSupportGz {
		contentType := w.Header().Get("Content-Type")

		if shouldCompress(contentType) {
			gz, err := gzip.NewWriterLevel(w.ResponseWriter, gzip.BestSpeed)
			if err == nil {
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Del("Content-Length")
				w.gz = gz
				w.writer = gz
			}
		}
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}
