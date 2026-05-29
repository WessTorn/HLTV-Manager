package api

import "net/http"

func (s *Server) handleStopHLTV(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := s.service.Stop(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "hltv stopped"})
}
