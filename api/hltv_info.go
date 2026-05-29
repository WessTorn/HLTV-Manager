package api

import (
	"net/http"
)

type HLTVInfoResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ShowIP     string `json:"show_ip"`
	Connect    string `json:"connect"`
	Port       string `json:"port"`
	GameID     string `json:"game_id"`
	Running    bool   `json:"running"`
	DemosCount int    `json:"demos_count"`
}

// @Summary     Get HLTV instance details
// @Description Returns one HLTV instance by ID with current runtime status and demos count.
// @Tags        hltv
// @Accept      json
// @Produce     json
// @Param       id path int true "HLTV ID"
// @Success     200 {object} HLTVInfoResponse "HLTV instance summary"
// @Failure     400 {object} ErrorResponse "Invalid HLTV ID"
// @Failure     404 {object} ErrorResponse "HLTV not found"
// @Router      /api/v1/hltv/{id} [get]
func (s *Server) handleGetHLTV(w http.ResponseWriter, id int) {
	s.mu.RLock()
	state, ok := s.states[id]
	if !ok {
		s.mu.RUnlock()
		writeError(w, http.StatusNotFound, "hltv not found")
		return
	}
	instance := state.Instance
	running := state.Running
	settings := state.Settings
	s.mu.RUnlock()

	demosCount := 0
	if instance != nil {
		_ = instance.DemoControl()
		demosCount = len(instance.SnapshotDemos())
	}

	writeJSON(w, http.StatusOK, HLTVInfoResponse{
		ID:         id,
		Name:       settings.Name,
		ShowIP:     settings.ShowIP,
		Connect:    settings.Connect,
		Port:       settings.Port,
		GameID:     settings.GameID,
		Running:    running,
		DemosCount: demosCount,
	})
}
