package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func main() {
	http.HandleFunc("/authenticate", BasicAuth(ProtectedHandler))
	http.ListenAndServe(":8080", nil)
}

// BasicAuth wraps a handler requiring HTTP basic authentication.
func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !checkCredentials(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

// checkCredentials checks the provided username and password against the expected values.
func checkCredentials(username, password string) bool {
	// Normally, you would use a more secure method for checking credentials.
	// Check if the username is an email from the "foo.com" domain.
	parts := strings.Split(username, "@")
	if len(parts) == 2 && parts[1] == "tm1-code.io" {
		// Only let the user in if he or she knows the magic password as well!
		return password == "apple"
	}
	return false
}

// ProtectedHandler is the handler for the protected endpoint.
func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user, _, _ := r.BasicAuth()
	response := struct {
		User string `json:"User"`
	}{
		User: user,
	}
	json.NewEncoder(w).Encode(response)
}
