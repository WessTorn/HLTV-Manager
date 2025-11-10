package main

import (
	"HLTV-Manager/config"
	"HLTV-Manager/hltv"
	log "HLTV-Manager/logger"
	"HLTV-Manager/reader"
	"HLTV-Manager/site"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	err := log.InitLogger("./log/")
	if err != nil {
		return
	}

	config.InitConfig()

	read, err := reader.ReadHLTVRunners()
	if err != nil {
		return
	}

	var hltvs []*hltv.HLTV

	for i, runner := range read {
		hltvConfig := hltv.Settings{
			Name:             runner.Name,
			ShowIP:           runner.ShowIP,
			Connect:          runner.Connect,
			Port:             runner.Port,
			GameID:           runner.GameID,
			DemoName:         runner.DemoName,
			MaxDemoDay:       runner.MaxDemoDay,
			DebugTerminalLog: runner.DebugTerminalLog,
			Cvars:            runner.Cvars,
		}

		h, err := hltv.NewHLTV(i+1, hltvConfig)
		if err != nil {
			continue
		}

			err = h.Start()
			if err != nil {
				continue
			}

			if err := h.DemoControl(); err != nil {
				log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Failed to preload demos: %v", h.ID, h.Settings.Name, err)
				continue
			}

			hltvs = append(hltvs, h)
		}

	site := &site.Site{HLTV: hltvs}
	go site.Init()

	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutDown

		for _, hltv := range hltvs {
			hltv.Quit()
		}

		time.Sleep(2 * time.Second)

		log.InfoLogger.Println("Программа завершена.")

		os.Exit(0)
	}()

	address := fmt.Sprintf("%s:%s", config.SiteIP(), config.SitePort())
	log.InfoLogger.Println("Starting site: ", address)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.ErrorLogger.Println("Server startup error: %v", err)
		shutDown <- syscall.SIGTERM
	}

	select {}
}

// for {
// 	in := bufio.NewReader(os.Stdin)
// 	line, err := in.ReadString('\n')
// 	if err != nil {
// 		fmt.Println("ERR")
// 		continue
// 	}

// 	err = hltv.WriteCommand(line)
// 	if err != nil {
// 		fmt.Println("Write to container error:", err)
// 		break
// 	}
// }
