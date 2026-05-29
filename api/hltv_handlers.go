package api

import (
	"HLTV-Manager/hltv"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (s *Server) hltvHandlers(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/hltv/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	parts := strings.Split(path, "/")

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid hltv id")
		return
	}

	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		s.handleGetHLTV(w, id)
		return
	case len(parts) == 2 && isHLTVAction(parts[1]):
		s.handleHLTVAction(w, r, id, parts[1])
		return
	case len(parts) == 2 && parts[1] == "demos" && r.Method == http.MethodGet:
		s.handleGetDemos(w, id)
		return
	case len(parts) == 4 && parts[1] == "demos" && parts[3] == "download" && r.Method == http.MethodGet:
		s.handleDownloadDemo(w, r, id, parts[2])
		return
	default:
		writeError(w, http.StatusNotFound, "route not found")
	}
}

func (s *Server) handleGetDemos(w http.ResponseWriter, id int) {
	s.mu.RLock()
	state, ok := s.states[id]
	if !ok {
		s.mu.RUnlock()
		writeError(w, http.StatusNotFound, "hltv not found")
		return
	}
	instance := state.Instance
	s.mu.RUnlock()

	if instance == nil {
		writeJSON(w, http.StatusOK, DemosResponse{Items: []hltv.Demos{}})
		return
	}

	_ = instance.DemoControl()
	writeJSON(w, http.StatusOK, DemosResponse{Items: instance.SnapshotDemos()})
}

func (s *Server) handleDownloadDemo(w http.ResponseWriter, r *http.Request, hltvID int, demoIDRaw string) {
	demoID, err := strconv.Atoi(demoIDRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid demo id")
		return
	}

	s.mu.RLock()
	state, ok := s.states[hltvID]
	if !ok {
		s.mu.RUnlock()
		writeError(w, http.StatusNotFound, "hltv not found")
		return
	}
	instance := state.Instance
	s.mu.RUnlock()

	if instance == nil {
		writeError(w, http.StatusBadRequest, "hltv was not started yet")
		return
	}

	demoName, demoPath, err := instance.GetDemoFile(demoID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid demo requested")
		return
	}

	if _, err := os.Stat(demoPath); os.IsNotExist(err) {
		writeError(w, http.StatusNotFound, "demo file not found")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+demoName)
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, demoPath)
}
