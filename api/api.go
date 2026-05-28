package api

import (
	"HLTV-Manager/hltv"
	"HLTV-Manager/reader"
	"net/http"
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

func (s *Server) InitAPI() {
	http.HandleFunc("/", withCORSFunc(s.rootHandler))
	http.HandleFunc("/api/v1/health", withCORSFunc(s.healthHandler))

	http.HandleFunc("/api/v1/hltv", withCORSFunc(s.hltvListHandler))

	http.HandleFunc("/api/v1/hltv/", withCORSFunc(s.hltvCommandsHandler))
}
