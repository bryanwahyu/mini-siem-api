package security

import (
    "net/http"
)

func APIKeyAuth(expected string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if expected == "" { next.ServeHTTP(w, r); return }
            if r.Header.Get("X-API-Key") != expected {
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte("{"+"\"error\":\"unauthorized\"}"))
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
