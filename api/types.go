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

type DemosResponse struct {
	Items []hltv.Demos `json:"items"`
}
