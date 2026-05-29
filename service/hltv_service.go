package service

import (
	"HLTV-Manager/hltv"
	log "HLTV-Manager/logger"
	"HLTV-Manager/repository"
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	ErrHLTVNotFound   = errors.New("hltv not found")
	ErrHLTVNotStarted = errors.New("hltv was not started yet")
	ErrInvalidDemo    = errors.New("invalid demo requested")
)

type Info struct {
	ID         int
	Name       string
	ShowIP     string
	Connect    string
	Port       string
	GameID     string
	Running    bool
	DemosCount int
}

type Service interface {
	StartAll()
	StopAll()
	Start(id int) error
	Stop(id int) error
	Restart(id int) error
	List() []Info
	Get(id int) (Info, error)
	GetDemos(id int) ([]hltv.Demos, error)
	GetDemoFile(id int, demoID int) (string, string, error)
}

type hltvState struct {
	id       int
	settings hltv.Settings
	instance *hltv.HLTV
	running  bool
}

type HLTVService struct {
	mu     sync.RWMutex
	states map[int]*hltvState
}

func NewHLTVService(repo repository.HLTVRepository) *HLTVService {
	states := make(map[int]*hltvState)
	for _, item := range repo.List() {
		states[item.ID] = &hltvState{
			id: item.ID,
			settings: hltv.Settings{
				Name:             item.Name,
				ShowIP:           item.ShowIP,
				Connect:          item.Connect,
				Port:             item.Port,
				GameID:           item.GameID,
				DemoName:         item.DemoName,
				MaxDemoDay:       item.MaxDemoDay,
				DebugTerminalLog: item.DebugTerminalLog,
				Cvars:            append([]string(nil), item.Cvars...),
			},
		}
	}

	return &HLTVService{states: states}
}

func (s *HLTVService) StartAll() {
	for _, id := range s.sortedIDs() {
		if err := s.Start(id); err != nil {
			log.WarningLogger.Printf("HLTV (ID: %d) failed to start on bootstrap: %v", id, err)
		}
	}
}

func (s *HLTVService) StopAll() {
	for _, id := range s.sortedIDs() {
		if err := s.Stop(id); err != nil {
			log.WarningLogger.Printf("HLTV (ID: %d) failed to stop on shutdown: %v", id, err)
		}
	}
}

func (s *HLTVService) Start(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[id]
	if !ok {
		return fmt.Errorf("hltv id %d not found", id)
	}

	if state.running {
		return fmt.Errorf("hltv id %d is already running", id)
	}

	if state.instance == nil {
		instance, err := hltv.NewHLTV(state.id, state.settings)
		if err != nil {
			return err
		}
		state.instance = instance
	}

	if err := state.instance.Start(); err != nil {
		return err
	}

	state.running = true
	if err := state.instance.DemoControl(); err != nil {
		log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Failed to preload demos: %v", state.id, state.settings.Name, err)
	}

	return nil
}

func (s *HLTVService) Stop(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[id]
	if !ok {
		return fmt.Errorf("hltv id %d not found", id)
	}

	if !state.running {
		return fmt.Errorf("hltv id %d is already stopped", id)
	}

	if state.instance == nil {
		return fmt.Errorf("hltv id %d has no initialized instance", id)
	}

	if err := state.instance.Quit(); err != nil {
		return err
	}

	state.running = false
	return nil
}

func (s *HLTVService) Restart(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.states[id]
	if !ok {
		return fmt.Errorf("hltv id %d not found", id)
	}

	if state.instance == nil {
		instance, err := hltv.NewHLTV(state.id, state.settings)
		if err != nil {
			return err
		}
		state.instance = instance
	}

	var err error
	if state.running {
		err = state.instance.Restart()
	} else {
		err = state.instance.Start()
	}
	if err != nil {
		return err
	}

	state.running = true
	if err := state.instance.DemoControl(); err != nil {
		log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Failed to preload demos: %v", state.id, state.settings.Name, err)
	}

	return nil
}

func (s *HLTVService) List() []Info {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]Info, 0, len(s.states))
	for _, id := range s.sortedIDsLocked() {
		state := s.states[id]

		demosCount := 0
		if state.instance != nil {
			demosCount = len(state.instance.SnapshotDemos())
		}

		list = append(list, Info{
			ID:         state.id,
			Name:       state.settings.Name,
			ShowIP:     state.settings.ShowIP,
			Connect:    state.settings.Connect,
			Port:       state.settings.Port,
			GameID:     state.settings.GameID,
			Running:    state.running,
			DemosCount: demosCount,
		})
	}

	return list
}

func (s *HLTVService) Get(id int) (Info, error) {
	s.mu.RLock()
	state, ok := s.states[id]
	if !ok {
		s.mu.RUnlock()
		return Info{}, ErrHLTVNotFound
	}
	instance := state.instance
	running := state.running
	settings := state.settings
	s.mu.RUnlock()

	demosCount := 0
	if instance != nil {
		_ = instance.DemoControl()
		demosCount = len(instance.SnapshotDemos())
	}

	return Info{
		ID:         id,
		Name:       settings.Name,
		ShowIP:     settings.ShowIP,
		Connect:    settings.Connect,
		Port:       settings.Port,
		GameID:     settings.GameID,
		Running:    running,
		DemosCount: demosCount,
	}, nil
}

func (s *HLTVService) GetDemos(id int) ([]hltv.Demos, error) {
	s.mu.RLock()
	state, ok := s.states[id]
	if !ok {
		s.mu.RUnlock()
		return nil, ErrHLTVNotFound
	}
	instance := state.instance
	s.mu.RUnlock()

	if instance == nil {
		return []hltv.Demos{}, nil
	}

	_ = instance.DemoControl()
	return instance.SnapshotDemos(), nil
}

func (s *HLTVService) GetDemoFile(id int, demoID int) (string, string, error) {
	s.mu.RLock()
	state, ok := s.states[id]
	if !ok {
		s.mu.RUnlock()
		return "", "", ErrHLTVNotFound
	}
	instance := state.instance
	s.mu.RUnlock()

	if instance == nil {
		return "", "", ErrHLTVNotStarted
	}

	demoName, demoPath, err := instance.GetDemoFile(demoID)
	if err != nil {
		return "", "", ErrInvalidDemo
	}

	return demoName, demoPath, nil
}

func (s *HLTVService) sortedIDs() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sortedIDsLocked()
}

func (s *HLTVService) sortedIDsLocked() []int {
	ids := make([]int, 0, len(s.states))
	for id := range s.states {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}
