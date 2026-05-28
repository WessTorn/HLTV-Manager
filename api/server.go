package api

import (
	"HLTV-Manager/hltv"
	log "HLTV-Manager/logger"
	"HLTV-Manager/reader"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type hltvState struct {
	ID       int
	Settings hltv.Settings
	Instance *hltv.HLTV
	Running  bool
}

type Server struct {
	mu     sync.RWMutex
	states map[int]*hltvState
}

type hltvSummaryResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ShowIP     string `json:"show_ip"`
	Connect    string `json:"connect"`
	Port       string `json:"port"`
	GameID     string `json:"game_id"`
	Running    bool   `json:"running"`
	DemosCount int    `json:"demos_count"`
}

type hltvDetailsResponse struct {
	hltvSummaryResponse
	Demos []hltv.Demos `json:"demos"`
}

func NewServer(runners []reader.HLTV) *Server {
	states := make(map[int]*hltvState, len(runners))

	for i, runner := range runners {
		id := i + 1
		states[id] = &hltvState{
			ID: id,
			Settings: hltv.Settings{
				Name:             runner.Name,
				ShowIP:           runner.ShowIP,
				Connect:          runner.Connect,
				Port:             runner.Port,
				GameID:           runner.GameID,
				DemoName:         runner.DemoName,
				MaxDemoDay:       runner.MaxDemoDay,
				DebugTerminalLog: runner.DebugTerminalLog,
				Cvars:            runner.Cvars,
			},
		}
	}

	return &Server{states: states}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.rootHandler)
	mux.HandleFunc("/api/v1/health", s.healthHandler)
	mux.HandleFunc("/api/v1/hltv", s.hltvRootHandler)
	mux.HandleFunc("/api/v1/hltv/", s.hltvByIDHandler)
	return withCORS(mux)
}

func (s *Server) StartAll() {
	for _, id := range s.sortedIDs() {
		if err := s.Start(id); err != nil {
			log.WarningLogger.Printf("HLTV (ID: %d) failed to start on bootstrap: %v", id, err)
		}
	}
}

func (s *Server) StopAll() {
	for _, id := range s.sortedIDs() {
		if err := s.Stop(id); err != nil {
			log.WarningLogger.Printf("HLTV (ID: %d) failed to stop on shutdown: %v", id, err)
		}
	}
}

func (s *Server) Start(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[id]
	if !ok {
		return fmt.Errorf("hltv id %d not found", id)
	}

	if state.Running {
		return fmt.Errorf("hltv id %d is already running", id)
	}

	if state.Instance == nil {
		instance, err := hltv.NewHLTV(state.ID, state.Settings)
		if err != nil {
			return err
		}
		state.Instance = instance
	}

	if err := state.Instance.Start(); err != nil {
		return err
	}

	state.Running = true
	if err := state.Instance.DemoControl(); err != nil {
		log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Failed to preload demos: %v", state.ID, state.Settings.Name, err)
	}

	return nil
}

func (s *Server) Stop(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[id]
	if !ok {
		return fmt.Errorf("hltv id %d not found", id)
	}

	if !state.Running {
		return fmt.Errorf("hltv id %d is already stopped", id)
	}

	if state.Instance == nil {
		return fmt.Errorf("hltv id %d has no initialized instance", id)
	}

	if err := state.Instance.Quit(); err != nil {
		return err
	}

	state.Running = false
	return nil
}

func (s *Server) Restart(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[id]
	if !ok {
		return fmt.Errorf("hltv id %d not found", id)
	}

	if state.Instance == nil {
		instance, err := hltv.NewHLTV(state.ID, state.Settings)
		if err != nil {
			return err
		}
		state.Instance = instance
	}

	var err error
	if state.Running {
		err = state.Instance.Restart()
	} else {
		err = state.Instance.Start()
	}

	if err != nil {
		return err
	}

	state.Running = true
	if err := state.Instance.DemoControl(); err != nil {
		log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Failed to preload demos: %v", state.ID, state.Settings.Name, err)
	}

	return nil
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "hltv-manager-backend",
		"api":     "/api/v1",
	})
}

func (s *Server) hltvRootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]hltvSummaryResponse, 0, len(s.states))
	for _, id := range s.sortedIDsLocked() {
		state := s.states[id]

		demosCount := 0
		if state.Instance != nil {
			demosCount = len(state.Instance.SnapshotDemos())
		}

		list = append(list, hltvSummaryResponse{
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

	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (s *Server) hltvByIDHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/hltv/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) >= 1 {
		switch parts[0] {
		case "start-all":
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			s.StartAll()
			writeJSON(w, http.StatusOK, map[string]string{"message": "all hltv instances were started"})
			return
		case "stop-all":
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			s.StopAll()
			writeJSON(w, http.StatusOK, map[string]string{"message": "all hltv instances were stopped"})
			return
		}
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid hltv id")
		return
	}

	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		s.handleGetHLTV(w, id)
		return
	case len(parts) == 2 && parts[1] == "start" && r.Method == http.MethodPost:
		if err := s.Start(id); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "hltv started"})
		return
	case len(parts) == 2 && parts[1] == "stop" && r.Method == http.MethodPost:
		if err := s.Stop(id); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "hltv stopped"})
		return
	case len(parts) == 2 && parts[1] == "restart" && r.Method == http.MethodPost:
		if err := s.Restart(id); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "hltv restarted"})
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

	var demos []hltv.Demos
	if instance != nil {
		_ = instance.DemoControl()
		demos = instance.SnapshotDemos()
	}

	response := hltvDetailsResponse{
		hltvSummaryResponse: hltvSummaryResponse{
			ID:         id,
			Name:       settings.Name,
			ShowIP:     settings.ShowIP,
			Connect:    settings.Connect,
			Port:       settings.Port,
			GameID:     settings.GameID,
			Running:    running,
			DemosCount: len(demos),
		},
		Demos: demos,
	}

	writeJSON(w, http.StatusOK, response)
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
		writeJSON(w, http.StatusOK, map[string]any{"items": []hltv.Demos{}})
		return
	}

	_ = instance.DemoControl()
	writeJSON(w, http.StatusOK, map[string]any{"items": instance.SnapshotDemos()})
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

func (s *Server) sortedIDs() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sortedIDsLocked()
}

func (s *Server) sortedIDsLocked() []int {
	ids := make([]int, 0, len(s.states))
	for id := range s.states {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error":   message,
		"status":  status,
		"success": false,
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
