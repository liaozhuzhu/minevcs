package main

import (
	"context"
	"drive/drive"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type ProgressReader struct {
	io.Reader
	Reporter func(bytesRead int64)
}

type UserData struct {
	MinecraftDirectory string `json:"minecraftDirectory"`
	WorldName          string `json:"worldName"`
}

type App struct {
	ctx context.Context
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.Reporter(int64(n))
	return
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	var minecraftWasRunning bool

	go func() {
		for {
			running, err := a.CheckMinecraftRunning()
			if err != nil {
				println("Error checking Minecraft status:", err.Error())
			} else if running {
				if !minecraftWasRunning {
					println("Minecraft is running ✅")
					minecraftWasRunning = true
					if authenticated, err := a.CheckIfAuthenticated(); err == nil && authenticated {
						a.pullWorld()
					}
				}
			} else {
				if minecraftWasRunning {
					println("Minecraft exited ❌")
				}
				minecraftWasRunning = false
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func (a *App) CheckMinecraftRunning() (bool, error) {
	path := "/Applications/Minecraft.app/Contents/MacOS/launcher"
	procs, err := process.Processes()
	if err != nil {
		return false, err
	}
	for _, p := range procs {
		exe, err := p.Exe()
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(exe), strings.ToLower(path)) {
			return true, nil
		}
	}
	return false, nil
}

func (a *App) GetWorlds(minecraftDirectory string) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(home, minecraftDirectory)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	var worlds []string
	for _, entry := range entries {
		if entry.Name() == ".DS_Store" {
			continue
		}
		worlds = append(worlds, entry.Name())
	}
	return worlds, nil
}

func (a *App) CloudUpload(worldName string, minecraftDirectory string) ([]string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	fullPath := home
	if worldName != "" {
		fullPath = filepath.Join(home, minecraftDirectory, worldName)
	}
	srv, err := drive.InitDrive()
	if err != nil {
		return nil, err
	}
	folderID, err := drive.UploadFolder(srv, fullPath, "")
	if err != nil {
		return nil, err
	}
	println("FINISHED UPLOADING WORLD: ", folderID)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return files, nil
}

func (a *App) GoogleAuth() (string, error) {
	url, err := drive.Authenticate()
	if err != nil {
		return "", err
	}
	return url, nil
}

func (a *App) UserAuthCode(code string) {
	drive.HandleAuthCode(code)
}

func (a *App) CheckIfAuthenticated() (bool, error) {
	_, err := os.Stat("token.json")
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *App) pullWorld() {
	println("Pulling world...")
}

func (a *App) SaveUserData(minecraftDirectory string, worldName string) {
	println("Saving user data locally")
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".minevcs")
	os.MkdirAll(configPath, os.ModePerm)
	file := filepath.Join(configPath, "config.json")
	data := fmt.Sprintf(`{"minecraftDirectory": "%s", "worldName": "%s", "lastUpdated": "%s"}`, minecraftDirectory, worldName, time.Now().Format(time.RFC3339))
	err := os.WriteFile(file, []byte(data), 0644)
	if err != nil {
		println("Error saving user data:", err.Error())
	} else {
		println("Uploading user data to drive")
	}
	// push this file to user's drive
	srv, err := drive.InitDrive()
	if err != nil {
		println("Error initializing drive:", err.Error())
		return
	}
	ctx := context.Background()
	configFile, err := os.Open(file)
	if err != nil {
		println("Error opening file:", err.Error())
		return
	}
	defer configFile.Close()
	createdFile, err := drive.UploadFile(ctx, srv, configFile, "")
	if err != nil {
		println("Error uploading file:", err.Error())
		return
	}
	println("User data uploaded successfully:", createdFile.Name)
}

func (a *App) GetUserData() (*UserData, error) {
	// first check if the config file exists locally (cached)
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".minevcs", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// if doesn't exist we need to pull it from the drive
			println("Config file not found locally, pulling from drive...")
			srv, err := drive.InitDrive()
			if err != nil {
				return nil, err
			}
			file, err := drive.FindFileByName(srv, "config.json")
			if err != nil {
				return nil, fmt.Errorf("failed to find config on Drive: %w", err)
			}

			ctx := context.Background()
			os.MkdirAll(filepath.Dir(configPath), 0755)
			err = drive.DownloadFile(ctx, srv, file.Id, configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to download config: %w", err)
			}

			println("Downloaded config from Drive ✅")

			data, err = os.ReadFile(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read downloaded config: %w", err)
			}

			// Use the data after reading the file
			println("Config data successfully read:", string(data))

		} else {
			return nil, err
		}
	}

	var config map[string]string
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &UserData{
		MinecraftDirectory: config["minecraftDirectory"],
		WorldName:          config["worldName"],
	}, nil
}
