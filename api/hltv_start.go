package api

import "net/http"

func (s *Server) handleStartHLTV(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := s.service.Start(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "hltv started"})
}
