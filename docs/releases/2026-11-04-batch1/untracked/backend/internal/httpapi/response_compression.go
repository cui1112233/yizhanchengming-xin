package httpapi

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strconv"
	"strings"
)

// minJSONCompressionBytes keeps tiny JSON responses (health checks, small
// errors) uncompressed: the gzip header would cost more than the payload.
const minJSONCompressionBytes = 1024

// withJSONCompression compresses JSON responses when the client accepts gzip.
// It is deliberately a thin wrapper around the whole router so it does not
// interact with ServeContent's Range/304 handling: image and other non-JSON
// responses are passed through untouched before any wrapping happens.
func withJSONCompression(next http.Handler) http.Handler {
	if next == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead || r.Header.Get("Range") != "" || !acceptsGzipEncoding(r) {
			next.ServeHTTP(w, r)
			return
		}
		writer := &compressingResponseWriter{ResponseWriter: w, threshold: minJSONCompressionBytes}
		next.ServeHTTP(writer, r)
		writer.finish()
	})
}

func acceptsGzipEncoding(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		name := token
		quality := 1.0
		if index := strings.IndexByte(token, ';'); index >= 0 {
			name = strings.TrimSpace(token[:index])
			for _, param := range strings.Split(token[index+1:], ";") {
				param = strings.TrimSpace(param)
				if strings.HasPrefix(param, "q=") {
					if parsed, err := strconv.ParseFloat(strings.TrimPrefix(param, "q="), 64); err == nil {
						quality = parsed
					}
				}
			}
		}
		if strings.EqualFold(name, "gzip") && quality > 0 {
			return true
		}
	}
	return false
}

func isJSONContentType(value string) bool {
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func appendVary(header http.Header, value string) {
	for _, existing := range header.Values("Vary") {
		for _, part := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}

// compressingResponseWriter delays committing the header until either the body
// grows past the threshold or the handler returns. That lets a small JSON body
// stay uncompressed while a large one starts streaming gzip immediately.
type compressingResponseWriter struct {
	http.ResponseWriter
	status      int
	compress    bool
	written     bool
	compressing bool
	buffer      bytes.Buffer
	gz          *gzip.Writer
	threshold   int
}

func (w *compressingResponseWriter) WriteHeader(status int) {
	// Informational responses are followed by a final status; pass them through.
	if status < 200 {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	if w.written {
		return
	}
	w.status = status
	header := w.Header()
	switch status {
	case http.StatusNoContent, http.StatusNotModified:
		w.compress = false
	default:
		w.compress = header.Get("Content-Encoding") == "" && isJSONContentType(header.Get("Content-Type"))
	}
	if w.compress {
		appendVary(header, "Accept-Encoding")
	}
	if !w.compress {
		w.ResponseWriter.WriteHeader(status)
		w.written = true
	}
}

func (w *compressingResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	if !w.compress {
		return w.ResponseWriter.Write(p)
	}
	if w.compressing {
		return w.gz.Write(p)
	}
	if _, err := w.buffer.Write(p); err != nil {
		return 0, err
	}
	if w.buffer.Len() >= w.threshold {
		if err := w.startCompressing(); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (w *compressingResponseWriter) Flush() {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	if w.compress {
		if !w.compressing {
			if err := w.startCompressing(); err != nil {
				return
			}
		}
		_ = w.gz.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (w *compressingResponseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *compressingResponseWriter) startCompressing() error {
	header := w.Header()
	header.Del("Content-Length")
	header.Set("Content-Encoding", "gzip")
	appendVary(header, "Accept-Encoding")
	w.ResponseWriter.WriteHeader(w.status)
	w.written = true
	w.compressing = true
	w.gz = gzip.NewWriter(w.ResponseWriter)
	if w.buffer.Len() > 0 {
		if _, err := w.buffer.WriteTo(w.gz); err != nil {
			return err
		}
		w.buffer.Reset()
	}
	return nil
}

func (w *compressingResponseWriter) finish() {
	if !w.compress {
		if !w.written {
			w.ResponseWriter.WriteHeader(w.statusOrOK())
			w.written = true
		}
		return
	}
	if !w.compressing {
		// The handler returned before reaching the threshold: keep the body
		// uncompressed, which also keeps Content-Length semantics intact.
		w.ResponseWriter.WriteHeader(w.statusOrOK())
		w.written = true
		if w.buffer.Len() > 0 {
			_, _ = w.ResponseWriter.Write(w.buffer.Bytes())
		}
		return
	}
	_ = w.gz.Close()
}

func (w *compressingResponseWriter) statusOrOK() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}
