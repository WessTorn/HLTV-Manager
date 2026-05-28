package api

import (
	"HLTV-Manager/hltv"
	log "HLTV-Manager/logger"
	"fmt"
	"sort"
)

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
