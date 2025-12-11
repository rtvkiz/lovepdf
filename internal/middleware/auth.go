package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// BasicAuth provides simple username/password authentication
func BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get credentials from environment variables
		expectedUsername := os.Getenv("AUTH_USERNAME")
		expectedPassword := os.Getenv("AUTH_PASSWORD")

		// Skip auth if not configured
		if expectedUsername == "" || expectedPassword == "" {
			next.ServeHTTP(w, r)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="LovePDF Tool"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Authentication required"))
			return
		}

		// Use constant time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(expectedUsername)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) == 1

		if !usernameMatch || !passwordMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="LovePDF Tool"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid credentials"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
