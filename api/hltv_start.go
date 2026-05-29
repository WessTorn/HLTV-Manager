package api

import "net/http"

// @Summary     Start HLTV instance
// @Description Starts one HLTV instance by ID.
// @Tags        hltv
// @Accept      json
// @Produce     json
// @Param       id path int true "HLTV ID"
// @Success     200 {object} MessageResponse "HLTV started"
// @Failure     400 {object} ErrorResponse "Start error"
// @Failure     405 {object} ErrorResponse "Method not allowed"
// @Router      /api/v1/hltv/{id}/start [post]
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
