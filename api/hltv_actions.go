package api

import "net/http"

type hltvAction struct {
	run     func(s *Server, id int) error
	message string
}

var hltvActions = map[string]hltvAction{
	"start": {
		run:     func(s *Server, id int) error { return s.Start(id) },
		message: "hltv started",
	},
	"stop": {
		run:     func(s *Server, id int) error { return s.Stop(id) },
		message: "hltv stopped",
	},
	"restart": {
		run:     func(s *Server, id int) error { return s.Restart(id) },
		message: "hltv restarted",
	},
}

func isHLTVAction(action string) bool {
	_, ok := hltvActions[action]
	return ok
}

func (s *Server) handleHLTVAction(w http.ResponseWriter, r *http.Request, id int, action string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cmd, ok := hltvActions[action]
	if !ok {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	if err := cmd.run(s, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: cmd.message})
}
