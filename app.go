package main

import (
	"archive/zip"
	"context"
	"crypto/sha256"
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
	logs               []string
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
	a.createMinevcsDirectory()
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".minevcs", "config.json")
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			a.printAndEmit("Config file not created yet, create a new one first" + " ❌")
			return
		}
		a.printAndEmit("Config file is corrupted, please create a new one ❌")
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		a.printAndEmit("Config file is corrupted, please create a new one ❌")
		return
	}

	var config map[string]string
	err = json.Unmarshal(data, &config)
	if err != nil {
		a.printAndEmit("Config file is corrupted, please create a new one ❌")
		return
	}

	a.minecraftLauncher = config["minecraftLauncher"]
	a.minecraftDirectory = config["minecraftDirectory"]
	a.worldName = config["worldName"]

	home, _ = os.UserHomeDir()
	worldPath := filepath.Join(home, a.minecraftDirectory, a.worldName)
	a.pushIfAhead(worldPath)

	if !a.isMonitoring {
		a.startMinecraftMonitor()
	}
}

func (a *App) pushIfAhead(worldPath string) {
	inSyncWithCloud, err := a.checkOutOfSync(worldPath)
	if err != nil {
		a.printAndEmit("Error checking world sync status: " + err.Error() + " ❌")
		return
	}
	if !inSyncWithCloud {
		a.printAndEmit("Local world is ahead of last uploaded world, pushing updated world to Drive ⏳")
		_, err = a.cloudUpload(a.worldName, a.minecraftDirectory)
		if err != nil {
			a.printAndEmit("Error uploading world: " + err.Error() + " ❌")
			return
		}
	} else {
		println("Local world is in sync with last uploaded world")
	}
}

func (a *App) checkOutOfSync(worldFolder string) (bool, error) {
	_, srv, err := drive.InitDrive()
	if err != nil {
		return false, err
	}
	var latest time.Time

	err = filepath.Walk(worldFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			modTime := info.ModTime()
			if modTime.After(latest) {
				latest = modTime
			}
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to walk through world folder: %w", err)
	}
	latest = latest.UTC()

	lastUploadTime, err := drive.GetLatestUploadTime(srv, a.worldName)
	if err != nil {
		return false, nil // here the file is not found on cloud, so we can assume that the local world is ahead of the last upload
	}
	parsedLastUploadTime, err := time.Parse(time.RFC3339, lastUploadTime)
	if err != nil {
		return false, fmt.Errorf("failed to parse last upload time: %w", err)
	}
	if latest.After(parsedLastUploadTime) {
		return false, nil // false means out of sync with the last upload on cloud
	}
	return true, nil // true means in sync with the last upload on cloud
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

	lockFilePath := filepath.Join(home, ".minevcs", "temp.lock")
	if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
		a.printAndEmit("Lock file not found, please restart the app")
		return nil, fmt.Errorf("lock file not found")
	}
	lockFile, err := os.Open(lockFilePath)
	if err != nil {
		return nil, err
	}

	// begin upload
	worldPath := filepath.Join(home, minecraftDirectory, worldName)
	// check if the world folder exists
	if _, err := os.Stat(worldPath); os.IsNotExist(err) {
		a.printAndEmit("World folder not found on local machine (most likely this is the device you are syncing to) ❌")
		return nil, fmt.Errorf("world folder not found")
	}
	// if so then we can begin the upload by creating a lock file
	tempLockFile, err := drive.UploadFile(ctx, srv, lockFile, "")
	if err != nil {
		return nil, err
	}

	// then zip + upload the world folder
	zipFilePath, err := a.zipFolder(worldPath)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(zipFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// print file name
	createdFile, err := drive.UploadFile(ctx, srv, file, "")
	if err != nil {
		return nil, err
	}
	a.printAndEmit("World uploaded successfully to Drive ✅")
	err = os.Remove(zipFilePath)
	if err != nil {
		return nil, err
	}
	err = drive.DeleteFile(srv, tempLockFile.Id)
	if err != nil {
		a.printAndEmit("Error deleting lock file: " + err.Error() + " ❌")
		return nil, err
	}
	return []string{createdFile.Name}, nil
}

func (a *App) GoogleAuth() (string, error) {
	url, err := drive.Authenticate()
	if err != nil {
		return "", err
	}
	return url, nil
}

func (a *App) UserAuthCode(code string) error {
	_, err := drive.VerifyAuthCode(code)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
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
	ctx, srv, err := drive.InitDrive()
	if err != nil {
		a.printAndEmit("Error initializing Drive: " + err.Error() + " ❌")
		return
	}
	_, err = drive.FindFileByName(srv, "temp.lock")
	if err == nil {
		a.printAndEmit("World upload in progress from another machine, please restart the app and try again soon ❌")
		return
	}
	a.printAndEmit("Downloading world from Drive... ⌛️")
	worldToDownload := a.worldName
	zipFile, err := drive.FindFileByName(srv, worldToDownload+".zip")
	if err != nil {
		a.printAndEmit("Error finding file: " + err.Error() + " (it may not exist yet)")
		return
	}
	if zipFile == nil {
		a.printAndEmit("No file found with the name: " + worldToDownload + ".zip ❌")
		return
	}
	zipFilePath := filepath.Join(os.TempDir(), zipFile.Name)
	err = drive.DownloadFile(ctx, srv, zipFile.Id, zipFilePath)
	if err != nil {
		a.printAndEmit("Error downloading file: " + err.Error() + " ❌")
		return
	}
	extractDir, err := a.unzipFolder(zipFilePath)
	if err != nil {
		a.printAndEmit("Error extracting zip file: " + err.Error() + " ❌")
		return
	}
	// move the extracted folder to the minecraft directory
	home, _ := os.UserHomeDir()
	minecraftPath := filepath.Join(home, a.minecraftDirectory)
	// check if the minecraft world already exists
	existingWorldPath := filepath.Join(minecraftPath, a.worldName)
	if _, err := os.Stat(existingWorldPath); err == nil {
		a.printAndEmit("World already exists, deleting existing world...")
		err = os.RemoveAll(existingWorldPath)
		if err != nil {
			a.printAndEmit("Error deleting existing world: " + err.Error() + " ❌")
			return
		}
		a.printAndEmit("Existing world deleted successfully ✅")
	}
	err = os.Rename(extractDir+"/"+a.worldName, filepath.Join(minecraftPath, a.worldName))
	if err != nil {
		a.printAndEmit("Error moving extracted folder: " + err.Error() + " ❌")
		return
	}
	a.printAndEmit("World pulled successfully from Drive ✅")
}

func (a *App) SaveUserData(minecraftLauncher string, minecraftDirectory string, worldName string) {
	a.printAndEmit("Saving user data locally...")
	a.minecraftLauncher = minecraftLauncher
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".minevcs")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		a.printAndEmit("Minevcs directory not found, please create a new one first")
		return
	}
	file := filepath.Join(configPath, "config.json")
	data := fmt.Sprintf(`{"minecraftLauncher": "%s", "minecraftDirectory": "%s", "worldName": "%s", "lastUpdated": "%s"}`, minecraftLauncher, minecraftDirectory, worldName, time.Now().Format(time.RFC3339))
	err := os.WriteFile(file, []byte(data), 0644)
	if err != nil {
		a.printAndEmit("Error saving user data: " + err.Error())
	} else {
		a.printAndEmit("User data saved successfully ✅")
	}
	a.minecraftLauncher = minecraftLauncher
	a.minecraftDirectory = minecraftDirectory
	a.worldName = worldName

	worldPath := filepath.Join(home, a.minecraftDirectory, a.worldName)
	a.pushIfAhead(worldPath)

	f, err := os.Open(worldPath)
	if err != nil {
		println("Error opening file:", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		println("Error copying file:", err)
	}

	fmt.Printf("Hash: %x", h.Sum(nil))

	if !a.isMonitoring {
		a.startMinecraftMonitor()
	}
}

func (a *App) CompareHash(hash1 string, hash2 string) bool {
	return hash1 == hash2
}

func (a *App) GetUserData() (*UserData, error) {

	return &UserData{
		MinecraftLauncher:  a.minecraftLauncher,
		MinecraftDirectory: a.minecraftDirectory,
		WorldName:          a.worldName,
	}, nil
}

func (a *App) startMinecraftMonitor() {
	_, err := a.CheckIfAuthenticated()
	if err != nil {
		a.printAndEmit("Please authenticate first ❌")
		return
	}
	a.printAndEmit("Monitoring Minecraft status... 👀")
	a.isMonitoring = true
	go func() {
		var minecraftWasRunning bool
		var cancelPushLoop context.CancelFunc

		for {
			running, err := a.CheckMinecraftRunning()
			if err != nil {
				a.printAndEmit("Error checking Minecraft status: " + err.Error() + " ❌")
			} else if running {
				if !minecraftWasRunning {
					a.printAndEmit("Minecraft is running ✅")
					minecraftWasRunning = true

					if authenticated, err := a.CheckIfAuthenticated(); err == nil && authenticated {
						a.pullWorld()
					}
					a.printAndEmit("Currently playing. Syncing world: " + a.worldName + " ⌛️")
				}
			} else {
				if minecraftWasRunning {
					a.printAndEmit("Minecraft exited ❌")

					if authenticated, err := a.CheckIfAuthenticated(); err == nil && authenticated {
						a.printAndEmit("User exited game, pushing world to Drive...")
						a.cloudUpload(a.worldName, a.minecraftDirectory)
					}
					if cancelPushLoop != nil {
						cancelPushLoop()
						cancelPushLoop = nil
					}
				} else {
					a.printAndEmit("Minecraft is not running ⌛️")
				}
				minecraftWasRunning = false
			}

			time.Sleep(2 * time.Second)
		}
	}()
}

func (a *App) createMinevcsDirectory() {
	home, _ := os.UserHomeDir()
	minevcsPath := filepath.Join(home, ".minevcs")
	if _, err := os.Stat(minevcsPath); os.IsNotExist(err) {
		err = os.MkdirAll(minevcsPath, os.ModePerm)
		if err != nil {
			a.printAndEmit("Error creating .minevcs directory: " + err.Error() + " ❌")
			return
		}
	}
	lockFilePath := filepath.Join(minevcsPath, "temp.lock")
	if _, err := os.Stat(lockFilePath); os.IsNotExist(err) {
		err = os.WriteFile(lockFilePath, []byte(""), 0644)
		if err != nil {
			a.printAndEmit("Error creating lock file: " + err.Error() + " ❌")
			return
		}
	}
	a.printAndEmit("Initialized Service successfully ✅")
}

func (a *App) printAndEmit(msg string) {
	timestamp := time.Now().Format("15:04:05")
	full := fmt.Sprintf("[%s] %s", timestamp, msg)
	a.logs = append(a.logs, full)
	println(msg)
	runtime.EventsEmit(a.ctx, "log", full)
}

func (a *App) flushLogs() {
	for _, log := range a.logs {
		runtime.EventsEmit(a.ctx, "log", log)
	}
}
