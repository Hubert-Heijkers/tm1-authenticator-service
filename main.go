package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

// checkCredentials checks the provided username and password against the expected values.
func checkCredentials(username, password string) bool {
	// This is where the actual implementation of the authentication logic goes. For this sample
	// we will simple expect an to be an e-mail address in the 'example.com' domain in combination
	// with the well known magic TM1 password;-!
	parts := strings.Split(username, "@")
	if len(parts) == 2 && parts[1] == "example.com" {
		return password == "apple"
	}
	return false
}

// BasicAuth wraps a handler requiring HTTP basic authentication.
func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !checkCredentials(user, pass) {
			// The content of the WWW-Authenticate header, returned if authentication failed, should
			// contain enough information for the client to be able to provide the required credentials
			// in the Authorization header on a subsequent request to authenticate.
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

// commandHandler is the handler for the HTTP request
func commandHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user, _, _ := r.BasicAuth()
	response := struct {
		User string `json:"Name"`
	}{
		User: user,
	}
	json.NewEncoder(w).Encode(response)
}

var elog debug.Log

type authenticatorService struct {
	port string
}

func (m *authenticatorService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	go runServer(m.port)
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			break loop
		default:
			elog.Error(1, string(c.Cmd))
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runWindowsService(name string, port string) {
	run := svc.Run
	elog.Info(1, "starting "+name+" service on port "+port)
	err := run(name, &authenticatorService{port})
	if err != nil {
		elog.Error(1, "service "+name+" failed: "+err.Error())
		return
	}
	elog.Info(1, "service "+name+" stopped")
}

func runServer(port string) {
	// Setup the HTTP server and route
	http.HandleFunc("/ActiveUser", BasicAuth(commandHandler))

	// Start listening to the user-defined or default port
	http.ListenAndServe(":"+port, nil)
}

func main() {
	// Define a command-line flag for the port, with a default value of 8080
	port := flag.String("port", "8080", "Port for the HTTP server to listen on")
	flag.Parse() // Parse the command-line flags

	// Determine if the service is begin started as a Windows service
	isWindowsService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if we are running in a windows service: %v", err)
	}

	// Start the Authenticator service
	if isWindowsService {
		elog, err = eventlog.Open("TM1-Authenticator-Service")
		if err != nil {
			return
		}
		defer elog.Close()
		runWindowsService("TM1-Authenticator-Service", *port)
	} else {
		fmt.Printf("Starting Authenticator service on port %s...\n", *port)
		elog = debug.New("TM1-Authenticator-Service")
		runServer(*port)
	}
}
