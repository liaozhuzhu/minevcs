package main

import (
	"context"
	"io"

	"os"
	"path/filepath"
	"quickstart/quickstart"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type ProgressReader struct {
	io.Reader
	Reporter func(bytesRead int64)
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.Reporter(int64(n))
	return
}

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) CheckMinecraftRunning(expectedPath string) (bool, error) {
	procs, err := process.Processes()
	if err != nil {
		return false, err
	}
	for _, p := range procs {
		exe, err := p.Exe()
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(exe), strings.ToLower(expectedPath)) {
			return true, nil
		}
	}
	return false, nil
}

func (a *App) CloudUpload(filePath string) ([]string, error) {
	// _, _ = a.CheckMinecraftRunning("/Applications/Minecraft.app/Contents/MacOS/launcher")

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	fullPath := home
	if filePath != "" {
		fullPath = filepath.Join(home, "Library/Application Support/minecraft/saves", filePath)
	}
	srv, err := quickstart.InitDrive()
	if err != nil {
		return nil, err
	}
	folderID, err := quickstart.UploadFolder(srv, fullPath, "")
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
	url, err := quickstart.Authenticate()
	if err != nil {
		return "", err
	}
	return url, nil
}

func (a *App) UserAuthCode(code string) {
	quickstart.HandleAuthCode(code)
}
