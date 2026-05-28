package api

import (
	"encoding/json"
	"net/http"
)

type HLTVListItem struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ShowIP     string `json:"show_ip"`
	Connect    string `json:"connect"`
	Port       string `json:"port"`
	GameID     string `json:"game_id"`
	Running    bool   `json:"running"`
	DemosCount int    `json:"demos_count"`
}

type HLTVListResponse struct {
	Items []HLTVListItem `json:"items"`
}

// @Summary     Get list of HLTV instances
// @Description Returns all configured HLTV instances with current runtime status and demo count.
// @Tags        hltv
// @Accept      json
// @Produce     json
// @Success     200 {object} HLTVListResponse "List of HLTV instances"
// @Failure     405 {object} ErrorResponse "Method not allowed"
// @Router      /api/v1/hltv [get]
func (s *Server) hltvListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Защищаем state
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]HLTVListItem, 0, len(s.states))
	for _, id := range s.sortedIDsLocked() {
		state := s.states[id]

		demosCount := 0
		if state.Instance != nil {
			demosCount = len(state.Instance.SnapshotDemos())
		}

		list = append(list, HLTVListItem{
			ID:         state.ID,
			Name:       state.Settings.Name,
			ShowIP:     state.Settings.ShowIP,
			Connect:    state.Settings.Connect,
			Port:       state.Settings.Port,
			GameID:     state.Settings.GameID,
			Running:    state.Running,
			DemosCount: demosCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HLTVListResponse{Items: list})
}
