package site

import (
	"HLTV-Manager/hltv"
	"net/http"
)

type Site struct {
	HLTV []*hltv.HLTV
}

func (site *Site) Init() {
	http.HandleFunc("/", site.homeHandler)
	http.HandleFunc("/demos/", site.demosHandler)
	http.HandleFunc("/download/", site.downloadHandler)

	http.HandleFunc("/admin/login", site.adminLoginHandler)
	http.HandleFunc("/admin/", site.adminAuth(site.adminHandler))
	http.HandleFunc("/admin/create", site.adminAuth(site.createHLTVHandler))
}
