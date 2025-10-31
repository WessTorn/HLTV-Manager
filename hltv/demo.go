package hltv

import (
	log "HLTV-Manager/logger"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (h *HLTV) DemoControl() error {
	if err := h.ArchiveCompletedDemos(); err != nil {
		return err
	}

	err := h.LoadDemosFromFolder()
	if err != nil {
		return err
	}

	sort.Slice(h.Demos, func(i, j int) bool {
		dateI, errI := time.Parse("2006.01.02 15:04", h.Demos[i].Date+" "+h.Demos[i].Time)
		if errI != nil {
			log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Ошибка парсинга для демки: %d %v", h.ID, h.Settings.Name, h.Demos[i].ID, errI)
			return false
		}

		dateJ, errJ := time.Parse("2006.01.02 15:04", h.Demos[j].Date+" "+h.Demos[j].Time)
		if errJ != nil {
			log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Ошибка парсинга для демки: %d %v", h.ID, h.Settings.Name, h.Demos[i].ID, errI)
			return false
		}

		return dateI.After(dateJ)
	})

	err = h.DeleteOldDemos()
	if err != nil {
		return err
	}

	return nil
}

func (h *HLTV) ArchiveCompletedDemos() error {
	entries, err := os.ReadDir(h.Settings.DemoDir)
	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to read demos directory: %v", h.ID, h.Settings.Name, err)
		return err
	}

	var demoEntries []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".dem") {
			demoEntries = append(demoEntries, entry)
		}
	}

	if len(demoEntries) <= 1 {
		return nil
	}

	sort.Slice(demoEntries, func(i, j int) bool {
		infoI, errI := demoEntries[i].Info()
		if errI != nil {
			return false
		}
		infoJ, errJ := demoEntries[j].Info()
		if errJ != nil {
			return true
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	for _, entry := range demoEntries[1:] {
		srcPath := filepath.Join(h.Settings.DemoDir, entry.Name())
		zipPath := srcPath + ".zip"

		if _, err := os.Stat(zipPath); err == nil {
			if removeErr := os.Remove(srcPath); removeErr != nil && !os.IsNotExist(removeErr) {
				log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Failed to remove already archived demo %s: %v", h.ID, h.Settings.Name, entry.Name(), removeErr)
			}
			continue
		}

		if err := archiveDemoFile(srcPath, zipPath); err != nil {
			log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to archive demo %s: %v", h.ID, h.Settings.Name, entry.Name(), err)
			continue
		}

		if err := os.Remove(srcPath); err != nil && !os.IsNotExist(err) {
			log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Failed to remove demo after archiving %s: %v", h.ID, h.Settings.Name, entry.Name(), err)
			continue
		}

		log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) Archived demo: %s.zip", h.ID, h.Settings.Name, entry.Name())
	}

	return nil
}

func archiveDemoFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	zipWriter := zip.NewWriter(dstFile)

	writer, err := zipWriter.Create(filepath.Base(srcPath))
	if err != nil {
		_ = zipWriter.Close()
		return err
	}

	if _, err := io.Copy(writer, srcFile); err != nil {
		_ = zipWriter.Close()
		return err
	}

	if err := zipWriter.Close(); err != nil {
		return err
	}

	return nil
}

func (h *HLTV) LoadDemosFromFolder() error {
	var demos []Demos

	var id int

	err := filepath.Walk(h.Settings.DemoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Error accessing file: %v", h.ID, h.Settings.Name, err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		name := info.Name()
		if !strings.HasSuffix(name, ".dem") && !strings.HasSuffix(name, ".zip") {
			return nil
		}

		if strings.HasSuffix(name, ".dem") {
			return nil
		}

		demo, err := parseDemoFilename(name)
		if err != nil {
			log.WarningLogger.Printf("HLTV (ID: %d, Name: %s) Error parsing file: %s, %v", h.ID, h.Settings.Name, name, err)
			return nil
		}

		id++
		demo.ID = id
		demo.Name = name
		demo.Path = path
		demo.Archived = strings.HasSuffix(name, ".zip")
		demos = append(demos, demo)

		return nil
	})

	if err != nil {
		log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to walk through folder: %v", h.ID, h.Settings.Name, err)
		return err
	}

	h.Demos = demos
	return nil
}

func (h *HLTV) DeleteOldDemos() error {
	now := time.Now()

	for _, demo := range h.Demos {
		demoDate, err := time.Parse("2006.01.02 15:04", demo.Date+" "+demo.Time)
		if err != nil {
			log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to parse date for demo %s: %v", h.ID, h.Settings.Name, demo.Name, err)
			return err
		}

		maxDemoDay, err := strconv.Atoi(h.Settings.MaxDemoDay)
		if err != nil {
			log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Error converting MaxDemoDay for demo %s: %v", h.ID, h.Settings.Name, demo.Name, err)
			return err
		}

		if now.Sub(demoDate).Hours() > float64(maxDemoDay*24) {
			if demo.Path == "" {
				continue
			}
			err := os.Remove(demo.Path)
			if err != nil {
				log.ErrorLogger.Printf("HLTV (ID: %d, Name: %s) Failed to remove old demo %s: %v", h.ID, h.Settings.Name, demo.Name, err)
				return err
			}

			log.InfoLogger.Printf("HLTV (ID: %d, Name: %s) Removed old demo: %s", h.ID, h.Settings.Name, demo.Name)
		}
	}

	return nil
}

func (h *HLTV) GetDemoFile(demoID int) (string, string, error) {
	for _, d := range h.Demos {
		if d.ID == demoID {
			if d.Path == "" {
				return "", "", fmt.Errorf("demo path is empty")
			}
			return d.Name, d.Path, nil
		}
	}

	return "", "", fmt.Errorf("demo with id %d not found", demoID)
}
