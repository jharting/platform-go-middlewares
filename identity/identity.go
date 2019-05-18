package identity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

type identityKey int

const key identityKey = iota

// Internal is the "internal" field of an XRHID
type Internal struct {
	OrgID string `json:"org_id"`
}

// XRHID is the "identity" pricipal object set by Cloud Platform 3scale
type XRHID struct {
	AccountNumber string   `json:"account_number"`
	Internal      Internal `json:"internal"`
}

func getErrorText(code int, reason string) string {
	return http.StatusText(code) + ": " + reason
}

func doError(w http.ResponseWriter, code int, reason string) {
	http.Error(w, getErrorText(code, reason), code)
}

// Get returns the identity struct from the context
func Get(ctx context.Context) XRHID {
	return ctx.Value(key).(XRHID)
}

// Identity extracts the X-Rh-Identity header and places the contents into the
// request context
func Identity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawHeaders := r.Header["X-Rh-Identity"]

		// must have an x-rh-id header
		if len(rawHeaders) != 1 {
			doError(w, 400, "missing x-rh-identity header")
			return
		}

		// must be able to base64 decode header
		idRaw, err := base64.StdEncoding.DecodeString(rawHeaders[0])
		if err != nil {
			doError(w, 400, "unable to b64 decode x-rh-identity header")
			return
		}

		var jsonData XRHID
		err = json.Unmarshal(idRaw, &jsonData)
		if err != nil {
			doError(w, 400, "x-rh-identity header is does not contain vaild JSON")
			return
		}

		if jsonData.AccountNumber == "" || jsonData.AccountNumber == "-1" {
			doError(w, 400, "x-rh-identity header has an invalid or missing account number")
			return
		}

		if jsonData.Internal.OrgID == "" {
			doError(w, 400, "x-rh-identity header has an invalid or missing org_id")
			return
		}

		ctx := context.WithValue(r.Context(), key, jsonData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
