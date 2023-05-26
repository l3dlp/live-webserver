package livews

import (
	"fmt"
	"net/http"
)

// The type `dynamicFileServerHandler` is a struct that implements the `ServeHTTP` method to serve
// files from a specified directory.
type dynamicFileServerHandler struct {
}

func (h *dynamicFileServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logBuilder.WriteString(fmt.Sprintf("Received request: %s %s\n", r.Method, r.URL))
	http.FileServer(http.Dir(folderEntry.Text)).ServeHTTP(w, r)
}

// This function is a middleware that logs incoming HTTP requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logBuilder.WriteString(fmt.Sprintf("Received request: %s %s\n", r.Method, r.URL))
		next.ServeHTTP(w, r)
	})
}
