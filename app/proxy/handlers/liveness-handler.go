package handlers

import (
	"net/http"

	"github.com/discue/go-syscall-gatekeeper/app/uroot"
	kit "github.com/stfsy/go-api-kit/server/handlers"
)

func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	gatekeeperEnforced := uroot.GetIsGatekeeperEnforced()
	if gatekeeperEnforced {
		kit.SendText(w, "Ok")
	} else {
		w.WriteHeader(503)
	}
}
