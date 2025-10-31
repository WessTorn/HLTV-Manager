package hltv

import (
	"HLTV-Manager/config"
	log "HLTV-Manager/logger"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func createDemosDir(hltv *HLTV) (string, error) {
	demoPath := filepath.Join(config.HltvDemosDir(), "demos", strconv.Itoa(hltv.ID), "cstrike")

	err := os.MkdirAll(demoPath, 0755)
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Error when creating a directory for demos: %v", hltv.ID, hltv.Settings.Name, err)
		return "", err
	}

	err = os.Chown(demoPath, 1000, 1000)
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Error chown directory for demo: %v", hltv.ID, hltv.Settings.Name, err)
		return "", err
	}

	return demoPath, nil
}

func createHltvCfg(hltv *HLTV) (string, error) {
	cfgPath := filepath.Join(config.HltvDemosDir(), "demos", strconv.Itoa(hltv.ID), "hltv.cfg")

	file, err := os.OpenFile(cfgPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Error when opening hltv.cfg: %v", hltv.ID, hltv.Settings.Name, err)
		return "", err
	}
	defer file.Close()

	fileContent := strings.Join(hltv.Settings.Cvars, "\n") + "\n"
	_, err = file.Write([]byte(fileContent))
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Error when writing to hltv.cfg: %v", hltv.ID, hltv.Settings.Name, err)
		return "", err
	}

	err = os.Chown(cfgPath, 1000, 1000)
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Error chown file for hltv.cfg: %v", hltv.ID, hltv.Settings.Name, err)
		return "", err
	}

	return cfgPath, nil
}

func parseDemoFilename(filename string) (Demos, error) {
	nameForParsing := filename
	if strings.HasSuffix(nameForParsing, ".zip") {
		nameForParsing = strings.TrimSuffix(nameForParsing, ".zip")
	}

	re := regexp.MustCompile(`^[^-]+-(\d{10})-(.+)\.dem$`)

	matches := re.FindStringSubmatch(nameForParsing)
	if matches == nil || len(matches) != 3 {
		return Demos{}, fmt.Errorf("incorrect file name format")
	}

	datetime := matches[1] // например: "2504281730"
	mapName := matches[2]  // например: "de_aztec"

	if _, err := strconv.Atoi(datetime); err != nil {
		return Demos{}, fmt.Errorf("invalid datetime format: not numeric")
	}

	yy := datetime[:2]
	mm := datetime[2:4]
	dd := datetime[4:6]
	hh := datetime[6:8]
	min := datetime[8:10]

	date := fmt.Sprintf("20%s.%s.%s", yy, mm, dd)
	time := fmt.Sprintf("%s:%s", hh, min)

	return Demos{
		Name: filename,
		Date: date,
		Time: time,
		Map:  mapName,
	}, nil
}
