package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

type simpleAuthenticatorService struct{}

func (m *simpleAuthenticatorService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	go runServer()
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			default:
				elog.Error(1, string(c.Cmd))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func main() {
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if isInteractive {
		elog = debug.New("SimpleAuthenticatorService")
		runServer()
	} else {
		elog, err = eventlog.Open("SimpleAuthenticatorService")
		if err != nil {
			return
		}
		defer elog.Close()
		runService("SimpleAuthenticatorService", false)
	}
}

func runService(name string, isDebug bool) {
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	elog.Info(1, "starting "+name+" service")
	err := run(name, &simpleAuthenticatorService{})
	if err != nil {
		elog.Error(1, "service "+name+" failed: "+err.Error())
		return
	}
	elog.Info(1, "service "+name+" stopped")
}

func runServer() {
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
