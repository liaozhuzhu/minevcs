package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
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

func (a *App) getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(a.ctx, tok)
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

func (a *App) UploadFile(localFilePath, driveFolder string) (string, error) {
	ctx := a.ctx
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return "", fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return "", fmt.Errorf("unable to parse config: %v", err)
	}
	client := a.getClient(config)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("unable to create Drive client: %v", err)
	}

	file, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %v", err)
	}
	defer file.Close()

	driveFile := &drive.File{
		Name: filepath.Base(localFilePath),
	}
	if driveFolder != "" {
		driveFile.Parents = []string{driveFolder}
	}

	fmt.Println("Uploading:", localFilePath)
	res, err := srv.Files.Create(driveFile).
		Media(file, googleapi.ChunkSize(1024*1024)).
		Fields("id, name").
		Do()
	if err != nil {
		return "", fmt.Errorf("Drive upload failed: %v", err)
	}

	fmt.Printf("File '%s' uploaded with ID: %s\n", res.Name, res.Id)
	return res.Id, nil
}

func (a *App) ListFiles(filePath string) ([]string, error) {
	_, _ = a.CheckMinecraftRunning("/Applications/Minecraft.app/Contents/MacOS/launcher")
	if _, err := a.UploadFile("test.txt", ""); err != nil {
		return nil, err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	fullPath := home
	if filePath != "" {
		fullPath = filepath.Join(home, "Library/Application Support/minecraft/saves", filePath)
	}

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
