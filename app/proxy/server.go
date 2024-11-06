package server

import (
	"net/http"

	"github.com/discue/go-syscall-gatekeeper/app/proxy/handlers"
	"github.com/stfsy/go-api-kit/server"
	kit "github.com/stfsy/go-api-kit/server/handlers"
)

var s *server.Server

func Start() error {
	s = server.NewServer(&server.ServerConfig{
		MuxCallback: func(mux *http.ServeMux) {
			mux.HandleFunc("GET /live", handlers.LivenessHandler)
			mux.HandleFunc("/live", kit.MethodNotAllowedHandler)

			mux.HandleFunc("GET /health", handlers.LivenessHandler)
			mux.HandleFunc("/health", kit.MethodNotAllowedHandler)
		},
		ListenCallback: func() {
			// do sth just after listen was called on the server instance and
			// just before the server starts serving requests
		},
		// port override is optional but can be used if you want to
		// define the port manually. If empty the value of env.PORT is used.
		PortOverride: "8888",
	})

	return s.Start()
}

func Stop() {
	s.Stop()
}
