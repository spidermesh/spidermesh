package proxy

import (
	"net/http"

	"github.com/vulcand/oxy/forward"
)

// Forward http request
func Forward(w http.ResponseWriter, req *http.Request) error {
	fwd, err := forward.New()
	if err != nil {
		return err
	}
	fwd.ServeHTTP(w, req)
	return nil
}
