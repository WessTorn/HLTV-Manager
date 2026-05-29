package main

import (
	"HLTV-Manager/api"
	"HLTV-Manager/config"
	log "HLTV-Manager/logger"
	"HLTV-Manager/reader"
	"HLTV-Manager/repository"
	"HLTV-Manager/service"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title           HLTV-Manager backend
// @version         beta1.0
// @description     API for HLTV-Manager.
// @host            0.0.0.0:3030
// @BasePath        /
// @schemes         http
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

	repo := repository.NewInMemoryHLTVRepository(read)
	hltvService := service.NewHLTVService(repo)

	server := api.NewServer(hltvService)
	server.InitAPI()
	hltvService.StartAll()

	address := fmt.Sprintf("%s:%s", config.SiteIP(), config.SitePort())
	httpServer := &http.Server{
		Addr: address,
	}

	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, syscall.SIGINT, syscall.SIGTERM)
	shutdownDone := make(chan struct{})

	go func() {
		<-shutDown

		hltvService.StopAll()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.ErrorLogger.Printf("Graceful shutdown error: %v", err)
		}

		log.InfoLogger.Println("Программа завершена.")
		close(shutdownDone)
	}()

	log.InfoLogger.Println("Starting API server: ", address)
	err = httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.ErrorLogger.Printf("Server startup error: %v", err)
		shutDown <- syscall.SIGTERM
	}

	<-shutdownDone
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

// For dev

// swag init -g main.go -o docs --parseInternal

// docker compose up --build -d
