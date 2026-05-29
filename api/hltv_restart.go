package api

import "net/http"

// @Summary     Restart HLTV instance
// @Description Restarts one HLTV instance by ID.
// @Tags        hltv
// @Accept      json
// @Produce     json
// @Param       id path int true "HLTV ID"
// @Success     200 {object} MessageResponse "HLTV restarted"
// @Failure     400 {object} ErrorResponse "Restart error"
// @Failure     405 {object} ErrorResponse "Method not allowed"
// @Router      /api/v1/hltv/{id}/restart [post]
func (s *Server) handleRestartHLTV(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := s.service.Restart(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "hltv restarted"})
}
