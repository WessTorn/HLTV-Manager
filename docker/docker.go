package docker

import (
	log "HLTV-Manager/logger"
	"context"
	"os"
	"strconv"
	"time"

	Config "HLTV-Manager/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

type HltvContainerConfig struct {
	Cmd      []string
	DemoPath string
	CfgPath  string
	Mounts   []mount.Mount
	Hltv     Hltv
}

type Docker struct {
	client *client.Client
	Attach types.HijackedResponse
}

type Hltv struct {
	ID     int
	Name   string
	GameID string
}

const (
	GAME_APPID_CSTRIKE      = 10
	GAME_APPID_TFC          = 20
	GAME_APPID_DOD          = 30
	GAME_APPID_DMC          = 40
	GAME_APPID_GEARBOX      = 50
	GAME_APPID_RICOCHET     = 60
	GAME_APPID_VALVE        = 70
	GAME_APPID_CZERO        = 80
	GAME_APPID_CZEROR       = 100
	GAME_APPID_BSHIFT       = 130
	GAME_APPID_CSTRIKE_BETA = 150
)

func NewDockerClient() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Docker{client: cli}, nil
}

func (docker *Docker) CreateAndStart(config HltvContainerConfig) error {
	ctx := context.Background()

	err := docker.StopContainerIfExists(config.Hltv)
	if err != nil {
		return err
	}

	gameID, err := strconv.Atoi(config.Hltv.GameID)
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to Atoi game id: %v", config.Hltv.ID, config.Hltv.Name, err)
		return err
	}

	path := getGamePath(gameID)

	resp, err := docker.client.ContainerCreate(ctx, &container.Config{
		Image:        Config.HltvDocker(),
		Cmd:          config.Cmd,
		Tty:          true,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Env: []string{
			"TZ=" + os.Getenv("TZ"), // прокидываем из hltv-manager
		},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: config.DemoPath,
				Target: path,
			},
			{
				Type:   mount.TypeBind,
				Source: config.CfgPath,
				Target: "/home/hltv/hltv.cfg",
			},
			{
				Type:     mount.TypeBind,
				Source:   "/etc/localtime",
				Target:   "/etc/localtime",
				ReadOnly: true,
			},
			{
				Type:     mount.TypeBind,
				Source:   "/etc/timezone",
				Target:   "/etc/timezone",
				ReadOnly: true,
			},
		},
		AutoRemove: true,
	}, nil, nil, "hltv_"+strconv.Itoa(config.Hltv.ID))
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to create container: %v", config.Hltv.ID, config.Hltv.Name, err)
		return err
	}

	docker.Attach, err = docker.client.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to get attach container: %v", config.Hltv.ID, config.Hltv.Name, err)
		return err
	}

	err = docker.client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to start container: %v", config.Hltv.ID, config.Hltv.Name, err)
		return err
	}

	return nil
}

func (docker *Docker) StopContainerIfExists(hltv Hltv) error {
	ctx := context.Background()

	containerName := "hltv_" + strconv.Itoa(hltv.ID)
	containers, err := docker.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to list containers for %s: %v", hltv.ID, hltv.Name, containerName, err)
		return err
	}

	log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) Attempting to stop and remove existing container: %s", hltv.ID, hltv.Name, containerName)

	for _, cont := range containers {
		if cont.Names[0] == "/"+containerName {

			err := docker.client.ContainerStop(ctx, cont.ID, container.StopOptions{})
			if err != nil {
				log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to stop container %s: %v", hltv.ID, hltv.Name, containerName, err)
				return err
			}

			tries := 5
			var removeErr error
			for i := 0; i < tries; i++ {
				_, err := docker.client.ContainerInspect(ctx, cont.ID)
				if err != nil {
					if client.IsErrNotFound(err) {
						log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) Container already removed.", hltv.ID, hltv.Name)
						break
					}
					log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to inspect container: %v", hltv.ID, hltv.Name, err)
					return err
				}

				removeErr = docker.client.ContainerRemove(ctx, cont.ID, container.RemoveOptions{Force: true})
				if removeErr == nil {
					log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) Successfully removed container: %s", hltv.ID, hltv.Name, containerName)
					break
				} else {
					log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Error removing container, retrying in 3 seconds... (Attempt %d/%d): %v", hltv.ID, hltv.Name, i+1, tries, err)
					time.Sleep(3 * time.Second)
				}
			}

			if removeErr != nil {
				log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to remove container after %d attempts: %v", hltv.ID, hltv.Name, tries, removeErr)
				return err
			}
		}
	}

	return nil
}

func getGamePath(gameID int) string {
	var gameDir string
	switch gameID {
	case GAME_APPID_CSTRIKE:
		gameDir = "cstrike"
	case GAME_APPID_TFC:
		gameDir = "tfc"
	case GAME_APPID_DOD:
		gameDir = "dod"
	case GAME_APPID_DMC:
		gameDir = "dmc"
	case GAME_APPID_GEARBOX:
		gameDir = "gearbox"
	case GAME_APPID_RICOCHET:
		gameDir = "ricochet"
	case GAME_APPID_VALVE:
		gameDir = "valve"
	case GAME_APPID_CZERO:
		gameDir = "czero"
	case GAME_APPID_CZEROR:
		gameDir = "czeror"
	case GAME_APPID_BSHIFT:
		gameDir = "bshift"
	case GAME_APPID_CSTRIKE_BETA:
		gameDir = "cstrike_beta"
	default:
		gameDir = "unknown"
	}
	return "/home/hltv/" + gameDir
}
