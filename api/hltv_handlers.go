package api

import (
	"HLTV-Manager/service"
	"errors"
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
	case len(parts) == 2 && parts[1] == "start":
		s.handleStartHLTV(w, r, id)
		return
	case len(parts) == 2 && parts[1] == "stop":
		s.handleStopHLTV(w, r, id)
		return
	case len(parts) == 2 && parts[1] == "restart":
		s.handleRestartHLTV(w, r, id)
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
	demos, err := s.service.GetDemos(id)
	if errors.Is(err, service.ErrHLTVNotFound) {
		writeError(w, http.StatusNotFound, "hltv not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, DemosResponse{Items: demos})
}

func (s *Server) handleDownloadDemo(w http.ResponseWriter, r *http.Request, hltvID int, demoIDRaw string) {
	demoID, err := strconv.Atoi(demoIDRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid demo id")
		return
	}

	demoName, demoPath, err := s.service.GetDemoFile(hltvID, demoID)
	if errors.Is(err, service.ErrHLTVNotFound) {
		writeError(w, http.StatusNotFound, "hltv not found")
		return
	}
	if errors.Is(err, service.ErrHLTVNotStarted) {
		writeError(w, http.StatusBadRequest, "hltv was not started yet")
		return
	}
	if errors.Is(err, service.ErrInvalidDemo) {
		writeError(w, http.StatusBadRequest, "invalid demo requested")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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
