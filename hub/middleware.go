package hub

import (
	"net/http"
	"strings"
)

// apiRedirectRouter is an http middleware. It accepts an http.Handler and
// returns a new http.Handler. This function adds to a default
// api call (/api/a-function) the current api version (/api/v1.0/a-function).
// This avoids the usage of http.Redirect and a second HTTP call to the
// redirected URL.
func (hub *Hub) apiRedirectRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		if strings.Contains(req.URL.Path, "api") {
			res := hub.apiMatch.Find([]byte(req.URL.String()))
			if len(res) == 0 {
				req.URL.Path = strings.Replace(req.URL.Path, "api", "api/v"+hub.apiVersion+"", 1)
			}
		}
		next.ServeHTTP(w, req)
	})
}
