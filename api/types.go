package api

import "HLTV-Manager/hltv"

type ErrorResponse struct {
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Success bool   `json:"success"`
}

type RootResponse struct {
	Service string `json:"service"`
	API     string `json:"api"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type HLTVSummaryResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ShowIP     string `json:"show_ip"`
	Connect    string `json:"connect"`
	Port       string `json:"port"`
	GameID     string `json:"game_id"`
	Running    bool   `json:"running"`
	DemosCount int    `json:"demos_count"`
}

type HLTVDetailsResponse struct {
	HLTVSummaryResponse
	Demos []hltv.Demos `json:"demos"`
}

type DemosResponse struct {
	Items []hltv.Demos `json:"items"`
}
