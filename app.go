package main

import (
	"archive/zip"
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
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProgressReader struct {
	io.Reader
	Reporter func(bytesRead int64)
}

type UserData struct {
	MinecraftLauncher  string `json:"minecraftLauncher"`
	MinecraftDirectory string `json:"minecraftDirectory"`
	WorldName          string `json:"worldName"`
}

type App struct {
	ctx                context.Context
	minecraftLauncher  string
	minecraftDirectory string
	worldName          string
	isMonitoring       bool
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

	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".minevcs", "config.json")
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			println("Config file not created yet, create a new one first")
			a.emitLog("Config file not created yet, create a new one first")
			return
		}
		println("Error checking config file:", err.Error())
		a.emitLog("Error checking config file: " + err.Error())
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		println("Error reading config file:", err.Error())
		a.emitLog("Error reading config file: " + err.Error())
		return
	}

	var config map[string]string
	err = json.Unmarshal(data, &config)
	if err != nil {
		println("Error parsing config file:", err.Error())
		a.emitLog("Error parsing config file: " + err.Error())
		return
	}

	a.minecraftLauncher = config["minecraftLauncher"]
	a.minecraftDirectory = config["minecraftDirectory"]
	a.worldName = config["worldName"]

	if !a.isMonitoring {
		a.startMinecraftMonitor()
	}
}

func (a *App) CheckMinecraftRunning() (bool, error) {
	path := a.minecraftLauncher
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

func (a *App) unzipFolder(zipFilePath string) (string, error) {
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return "", err
	}
	defer zipReader.Close()

	extractDir := strings.TrimSuffix(zipFilePath, ".zip")
	err = os.MkdirAll(extractDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	for _, file := range zipReader.File {
		filePath := filepath.Join(extractDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		srcFile, err := file.Open()
		if err != nil {
			return "", err
		}
		defer srcFile.Close()

		destFile, err := os.Create(filePath)
		if err != nil {
			return "", err
		}
		defer destFile.Close()

		if _, err = io.Copy(destFile, srcFile); err != nil {
			return "", err
		}
	}

	return extractDir, nil
}

func (a *App) zipFolder(sourceDir string) (string, error) {
	// Compute output zip path
	parentDir := filepath.Dir(sourceDir)
	baseName := filepath.Base(sourceDir)
	zipFilePath := filepath.Join(parentDir, baseName+".zip")

	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, filepath.Dir(sourceDir))
		relPath = strings.TrimPrefix(relPath, string(filepath.Separator))

		if relPath == "" {
			return nil
		}

		if info.IsDir() {
			_, err = zipWriter.Create(relPath + "/")
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return zipFilePath, err
}

func (a *App) cloudUpload(worldName string, minecraftDirectory string) ([]string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	ctx, srv, err := drive.InitDrive()
	if err != nil {
		return nil, err
	}
	// zip the world folder
	worldPath := filepath.Join(home, minecraftDirectory, worldName)
	zipFilePath, err := a.zipFolder(worldPath)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(zipFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	createdFile, err := drive.UploadFile(ctx, srv, file, "")
	if err != nil {
		return nil, err
	}
	fmt.Println("Uploaded file:", createdFile.Name)
	a.emitLog("Uploaded file: " + createdFile.Name)
	err = os.Remove(zipFilePath)
	if err != nil {
		return nil, err
	}
	println("Deleted zip file:", zipFilePath)
	a.emitLog("Deleted zip file: " + zipFilePath)
	return []string{createdFile.Name}, nil
}

func (a *App) GoogleAuth() (string, error) {
	url, err := drive.Authenticate()
	if err != nil {
		return "", err
	}
	return url, nil
}

func (a *App) UserAuthCode(code string) {
	drive.VerifyAuthCode(code)
}

func (a *App) CheckIfAuthenticated() (bool, error) {
	home, _ := os.UserHomeDir()
	pathToToken := filepath.Join(home, ".minevcs", "token.json")
	_, err := os.Stat(pathToToken)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *App) pullWorld() {
	println("Downloading world from Drive...")
	a.emitLog("Downloading world from Drive...")
	ctx, srv, err := drive.InitDrive()
	if err != nil {
		println("Error initializing Drive:", err.Error())
		a.emitLog("Error initializing Drive: " + err.Error())
		return
	}
	worldToDownload := a.worldName
	zipFile, err := drive.FindFileByName(srv, worldToDownload+".zip")
	if err != nil {
		println("Error finding file:", err.Error())
		a.emitLog("Error finding file: " + err.Error())
		return
	}
	if zipFile == nil {
		println("No file found with the name:", worldToDownload+".zip")
		a.emitLog("No file found with the name: " + worldToDownload + ".zip")
		return
	}
	zipFilePath := filepath.Join(os.TempDir(), zipFile.Name)
	err = drive.DownloadFile(ctx, srv, zipFile.Id, zipFilePath)
	if err != nil {
		println("Error downloading file:", err.Error())
		a.emitLog("Error downloading file: " + err.Error())
		return
	}
	println("World downloaded successfully")
	a.emitLog("World downloaded successfully")
	extractDir, err := a.unzipFolder(zipFilePath)
	if err != nil {
		println("Error extracting zip file:", err.Error())
		a.emitLog("Error extracting zip file: " + err.Error())
		return
	}
	println("World extracted successfully to:", extractDir)
	a.emitLog("World extracted successfully to: " + extractDir)
	// move the extracted folder to the minecraft directory
	home, _ := os.UserHomeDir()
	minecraftPath := filepath.Join(home, a.minecraftDirectory)
	// check if the minecraft world already exists
	existingWorldPath := filepath.Join(minecraftPath, a.worldName)
	if _, err := os.Stat(existingWorldPath); err == nil {
		println("World already exists, deleting existing world...")
		a.emitLog("World already exists, deleting existing world...")
		err = os.RemoveAll(existingWorldPath)
		if err != nil {
			println("Error deleting existing world:", err.Error())
			a.emitLog("Error deleting existing world: " + err.Error())
			return
		}
		println("Existing world deleted successfully ✅")
		a.emitLog("Existing world deleted successfully ✅")
	}
	// move the extracted folder to the minecraft directory
	println("Writing world folder to:", minecraftPath)
	a.emitLog("Writing world folder to: " + minecraftPath)
	err = os.Rename(extractDir+"/"+a.worldName, filepath.Join(minecraftPath, a.worldName))
	if err != nil {
		println("Error moving extracted folder:", err.Error())
		a.emitLog("Error moving extracted folder: " + err.Error())
		return
	}
	println("World moved successfully to:", minecraftPath+"✅")
	a.emitLog("World moved successfully to: " + minecraftPath + "✅")
}

func (a *App) SaveUserData(minecraftLauncher string, minecraftDirectory string, worldName string) {
	println("Saving user data locally")
	a.minecraftLauncher = minecraftLauncher
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".minevcs")
	os.MkdirAll(configPath, os.ModePerm)
	file := filepath.Join(configPath, "config.json")
	data := fmt.Sprintf(`{"minecraftLauncher": "%s", "minecraftDirectory": "%s", "worldName": "%s", "lastUpdated": "%s"}`, minecraftLauncher, minecraftDirectory, worldName, time.Now().Format(time.RFC3339))
	err := os.WriteFile(file, []byte(data), 0644)
	if err != nil {
		println("Error saving user data:", err.Error())
		a.emitLog("Error saving user data: " + err.Error())
	} else {
		println("User data saved successfully ✅")
		a.emitLog("User data saved successfully ✅")
	}
	println("Config file path:", file)
	a.emitLog("Config file path: " + file)
	a.minecraftLauncher = minecraftLauncher
	a.minecraftDirectory = minecraftDirectory
	a.worldName = worldName
	if !a.isMonitoring {
		a.startMinecraftMonitor()
	}
}

func (a *App) GetUserData() (*UserData, error) {
	// first check if the config file exists locally (cached)
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".minevcs", "config.json")
	_, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			println("Config file not created yet, create a new one first")
			a.emitLog("Config file not created yet, create a new one first")
		}
		return nil, err
	}

	// var config map[string]string
	// err = json.Unmarshal(data, &config)
	// if err != nil {
	// 	return nil, err
	// }

	// a.minecraftLauncher = config["minecraftLauncher"]
	// a.minecraftDirectory = config["minecraftDirectory"]
	// a.worldName = config["worldName"]

	return &UserData{
		MinecraftLauncher:  a.minecraftLauncher,
		MinecraftDirectory: a.minecraftDirectory,
		WorldName:          a.worldName,
	}, nil
}

func (a *App) emitLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	runtime.EventsEmit(a.ctx, "log", fmt.Sprintf("[%s] %s", timestamp, message))
}

func (a *App) startMinecraftMonitor() {
	println("Starting Minecraft monitor... 👀")
	a.emitLog("Starting Minecraft monitor... 👀")
	a.isMonitoring = true
	go func() {
		var minecraftWasRunning bool
		var cancelPushLoop context.CancelFunc

		for {
			running, err := a.CheckMinecraftRunning()
			if err != nil {
				println("Error checking Minecraft status:", err.Error())
			} else if running {
				if !minecraftWasRunning {
					println("Minecraft is running ✅")
					a.emitLog("Minecraft is running ✅")
					minecraftWasRunning = true

					if authenticated, err := a.CheckIfAuthenticated(); err == nil && authenticated {
						a.pullWorld()
					}
				}
			} else {
				if minecraftWasRunning {
					println("Minecraft exited ❌")
					a.emitLog("Minecraft exited ❌")

					if authenticated, err := a.CheckIfAuthenticated(); err == nil && authenticated {
						println("User exited game, pushing world to Drive...")
						a.emitLog("User exited game, pushing world to Drive...")
						a.cloudUpload(a.worldName, a.minecraftDirectory)
					}
					if cancelPushLoop != nil {
						cancelPushLoop()
						cancelPushLoop = nil
					}
				} else {
					a.emitLog("Minecraft is not running ⏰")
				}
				minecraftWasRunning = false
			}

			time.Sleep(2 * time.Second)
		}
	}()
}
