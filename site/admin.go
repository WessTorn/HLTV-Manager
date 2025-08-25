package site

import (
	"HLTV-Manager/hltv"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
)

func (site *Site) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles(
			filepath.Join("frontend", "head.gohtml"),
			filepath.Join("frontend", "login.gohtml"),
		)
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(w, "login", nil) // <- обязательно имя шаблона
		if err != nil {
			http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Проверка логина
		user := r.FormValue("username")
		pass := r.FormValue("password")

		if user == "admin" && pass == "1234" {
			// Устанавливаем cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "admin",
				Value: "true",
				Path:  "/",
			})
			http.Redirect(w, r, "/admin/", http.StatusSeeOther)
			return
		}

		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
	}
}

func (site *Site) adminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin")
		if err != nil || cookie.Value != "true" {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (site *Site) adminHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"frontend/head.gohtml",
		"frontend/admin.gohtml",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = tmpl.ExecuteTemplate(w, "admin", site.HLTV)
}

type HLTVForm struct {
	Name       string
	ShowIP     string
	Connect    string
	Port       string
	GameID     string
	DemoName   string
	MaxDemoDay string
	Debug      bool
	Cvars      string
}

func (site *Site) createHLTVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles(
			filepath.Join("frontend", "head.gohtml"),
			filepath.Join("frontend", "create.gohtml"),
		)
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		example := HLTVForm{
			Name:       "HltvServer",
			ShowIP:     "85.192.30.229:27015",
			Connect:    "194.87.226.186:27015",
			Port:       "28015",
			GameID:     "10",
			DemoName:   "hltv2",
			MaxDemoDay: "3",
			Debug:      true,
			Cvars: `hostname "HLTV | Manager"
name "HLTV-Manager"
maxclients "0"
serverpassword "2332"
nomaster "1"
publicgame "0"
autoretry "1"
rate "100000"
updaterate "40"
maxrate "10000"
delay "0"
blockvoice "0"
signoncommands "voice_scale 2; voice_overdrive 16; volume 0.5"
chatmode "0"
logfile "0"`,
		}
		_ = tmpl.ExecuteTemplate(w, "create", example)
		return
	}

	if r.Method == http.MethodPost {
		newHLTV := &hltv.HLTV{
			ID: len(site.HLTV) + 1,
			Settings: hltv.Settings{
				Name:             r.FormValue("name"),
				ShowIP:           r.FormValue("showip"),
				Connect:          r.FormValue("connect"),
				Port:             r.FormValue("port"),
				GameID:           r.FormValue("gameid"),
				DemoName:         r.FormValue("demoname"),
				MaxDemoDay:       r.FormValue("maxdemoday"),
				DebugTerminalLog: r.FormValue("debug") == "on",
				Cvars:            strings.Split(strings.TrimSpace(r.FormValue("cvars")), "\n"),
			},
		}

		site.HLTV = append(site.HLTV, newHLTV)

		http.Redirect(w, r, "/admin/", http.StatusSeeOther)
	}
}
