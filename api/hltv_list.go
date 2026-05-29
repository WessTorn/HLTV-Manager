package api

import (
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
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	items := s.service.List()
	list := make([]HLTVListItem, 0, len(items))
	for _, state := range items {
		list = append(list, HLTVListItem{
			ID:         state.ID,
			Name:       state.Name,
			ShowIP:     state.ShowIP,
			Connect:    state.Connect,
			Port:       state.Port,
			GameID:     state.GameID,
			Running:    state.Running,
			DemosCount: state.DemosCount,
		})
	}

	writeJSON(w, http.StatusOK, HLTVListResponse{Items: list})
}
