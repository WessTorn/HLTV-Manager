package repository

import (
	"HLTV-Manager/reader"
	"sort"
	"sync"
)

type HLTVConfig struct {
	ID               int
	Name             string
	ShowIP           string
	Connect          string
	Port             string
	GameID           string
	DemoName         string
	MaxDemoDay       string
	DebugTerminalLog bool
	Cvars            []string
}

type HLTVRepository interface {
	List() []HLTVConfig
	Get(id int) (HLTVConfig, bool)
}

type InMemoryHLTVRepository struct {
	mu    sync.RWMutex
	items map[int]HLTVConfig
}

func NewInMemoryHLTVRepository(runners []reader.HLTV) *InMemoryHLTVRepository {
	items := make(map[int]HLTVConfig, len(runners))

	for i, runner := range runners {
		id := i + 1
		items[id] = HLTVConfig{
			ID:               id,
			Name:             runner.Name,
			ShowIP:           runner.ShowIP,
			Connect:          runner.Connect,
			Port:             runner.Port,
			GameID:           runner.GameID,
			DemoName:         runner.DemoName,
			MaxDemoDay:       runner.MaxDemoDay,
			DebugTerminalLog: runner.DebugTerminalLog,
			Cvars:            append([]string(nil), runner.Cvars...),
		}
	}

	return &InMemoryHLTVRepository{items: items}
}

func (r *InMemoryHLTVRepository) List() []HLTVConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]int, 0, len(r.items))
	for id := range r.items {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	result := make([]HLTVConfig, 0, len(ids))
	for _, id := range ids {
		item := r.items[id]
		item.Cvars = append([]string(nil), item.Cvars...)
		result = append(result, item)
	}

	return result
}

func (r *InMemoryHLTVRepository) Get(id int) (HLTVConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[id]
	if !ok {
		return HLTVConfig{}, false
	}

	item.Cvars = append([]string(nil), item.Cvars...)
	return item, true
}
