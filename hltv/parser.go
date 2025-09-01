package hltv

import (
	log "HLTV-Manager/logger"
	"fmt"
	"regexp"
	"strings"
)

type Status int

const (
	HLTV_CHECK_INIT   Status = iota // 0 HLTV инициализируется
	HLTV_FAIL_INIT                  // 1 HLTV иницилизирован
	HLTV_CONNECT                    // 2 HLTV подключается к серверу
	HLTV_FAIL_CONNECT               // 3 У HLTV проблеммы с подключением к серверу
	HLTV_RECORD                     // 4 HLTV Записывает демо
	HLTV_GOOD                       // 5 HLTV Все хорошо
)

type Parser struct {
	Status      Status
	initModules map[string]bool
}

var patterns = map[string]*regexp.Regexp{
	"timeout":      regexp.MustCompile(`^WARNING! Server::Challenge: Timeout after \d+ retries$`),
	"build":        regexp.MustCompile(`^BUILD \d+ SERVER \(\d+ CRC\)$`),
	"recording":    regexp.MustCompile(`^Start recording to [a-zA-Z0-9]+-\d+-[a-zA-Z0-9_]+\.dem.$`),
	"rejected":     regexp.MustCompile(`^Connection rejected: No password set.*$`),
	"disconnected": regexp.MustCompile(`^Disconnected.*$`),
	"server_state": regexp.MustCompile(`^Server::SetState: not valid m_ServerState \(\d+ -> \d+\).$`),
}

func (hltv *HLTV) TerminalControl() {
	hltv.Parser.Status = HLTV_CHECK_INIT
	hltv.Parser.initModules = map[string]bool{
		"Console initialized.":       false,
		"FileSystem initialized.":    false,
		"Network initialized.":       false,
		"Master module initialized.": false,
		"Server module initialized.": false,
		"World module initialized.":  false,
		"Demo client initialized.":   false,
		"Executing file hltv.cfg.":   false,
		"Proxy module initialized.":  false,
		"Recording initialized.":     false,
	}

	buf := make([]byte, 1024)
	for {
		n, err := hltv.Docker.Attach.Reader.Read(buf)
		if err != nil {
			break
		}
		line := string(buf[:n])
		line = strings.TrimSpace(line)
		hltv.ParseHltvOutLines(line)
	}
}

func (hltv *HLTV) ParseHltvOutLines(input string) {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.ReplaceAll(line, "\r", "")
		if line == "" {
			continue
		}

		if hltv.Settings.DebugTerminalLog {
			log.InfoLogger.Printf("(HLTV %d) %s", hltv.ID, line)
		}

		switch hltv.Parser.Status {
		case HLTV_CHECK_INIT:
			if initialized, exists := hltv.Parser.initModules[line]; exists {
				if !initialized {
					hltv.Parser.initModules[line] = true
				}
			}

			if hltv.allModulesInitialized() {
				hltv.Parser.Status = HLTV_CONNECT
				log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) All modules are initialized.", hltv.ID, hltv.Settings.Name)
			}
			continue
		case HLTV_CONNECT, HLTV_FAIL_CONNECT:
			{
				switch {
				case patterns["timeout"].MatchString(line):
					log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Cannot connect to server: %s", hltv.ID, hltv.Settings.Name, hltv.Settings.Connect)
				case patterns["rejected"].MatchString(line):
					log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Cannot connect to server due to password: %s", hltv.ID, hltv.Settings.Name, hltv.Settings.Connect)
				case patterns["build"].MatchString(line):
					log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) Connected to the server: %s", hltv.ID, hltv.Settings.Name, hltv.Settings.Connect)
					hltv.Parser.Status = HLTV_RECORD
				}
				continue
			}
		case HLTV_RECORD:
			{
				if patterns["recording"].MatchString(line) {
					log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) Start recorded demo.", hltv.ID, hltv.Settings.Name)
					fmt.Println("Началась запись демки:", line)
					hltv.Parser.Status = HLTV_GOOD
					hltv.DemoControl()
				}
				continue
			}
		case HLTV_GOOD:
			{
				switch {
				case patterns["build"].MatchString(line):
					hltv.Parser.Status = HLTV_RECORD
					log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) Reconnect.", hltv.ID, hltv.Settings.Name)
				case patterns["disconnected"].MatchString(line):
					hltv.Parser.Status = HLTV_CONNECT
					log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) Disconnected from server: %s", hltv.ID, hltv.Settings.Name, hltv.Settings.Connect)
				case patterns["server_state"].MatchString(line):
					log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Invalid server state: %s", hltv.ID, hltv.Settings.Name, line)
					_ = hltv.Restart()
					hltv.Parser.Status = HLTV_CHECK_INIT
				}
				continue
			}
		}

		// fmt.Printf("[%s]\n", line)
	}
}

func (hltv *HLTV) allModulesInitialized() bool {
	for _, initialized := range hltv.Parser.initModules {
		if !initialized {
			return false
		}
	}
	return true
}
