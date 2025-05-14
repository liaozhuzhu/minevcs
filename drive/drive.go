package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

//go:embed assets/credentials.json
var credentialsJSON []byte

// built off of quickstart example from google drive api w/ go (adjusted for UX)
func getClient(config *oauth2.Config) *http.Client {
	// local token file should already exist by now
	tok, err := tokenFromFile()
	if err != nil {
		println(err)
	}
	return config.Client(context.Background(), tok)
}

func getURL(config *oauth2.Config) string {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	_, err := tokenFromFile()
	if err != nil {
		// 🔥 Force consent to always get a refresh_token
		authURL := config.AuthCodeURL(
			"state-token",
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("prompt", "consent"),
		)
		return authURL
	}
	return ""
}

func VerifyAuthCode(code string) (*http.Client, error) {
	b := credentialsJSON
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}
	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, fmt.Errorf("Unable to verify token: %v", err)
	}
	tokFile := "token.json"
	saveToken(tokFile, tok)
	return config.Client(context.Background(), tok), nil
}

// Retrieves a token from a local file.
func tokenFromFile() (*oauth2.Token, error) {
	home, _ := os.UserHomeDir()
	file := home + "/.minevcs/token.json"
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(tokenFile string, token *oauth2.Token) {
	home, _ := os.UserHomeDir()
	// create a .minevcs folder if it doesn't exist
	if _, err := os.Stat(home + "/.minevcs"); os.IsNotExist(err) {
		err = os.Mkdir(home+"/.minevcs", 0700)
		if err != nil {
			log.Fatalf("Unable to create .minevcs folder: %v", err)
		}
	}
	tokenSavePath := home + "/.minevcs/" + tokenFile
	fmt.Printf("Saving token file to: %s\n", tokenSavePath)
	f, err := os.OpenFile(tokenSavePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// BEGIN GOOGLE DRIVE API
func UploadFile(ctx context.Context, srv *drive.Service, file *os.File, parentId string) (*drive.File, error) {
	var parents []string
	if parentId != "" {
		parents = []string{parentId}
	}
	fileName := filepath.Base(file.Name())
	println("File name:", fileName)
	// check if file already exists on drive if so delete it
	existingFile, err := FindFileByName(srv, fileName)
	if err != nil {
		println("File not found on drive, creating new file...")
	} else {
		println("File already exists on drive, deleting existing file...")
		err = DeleteFile(srv, existingFile.Id)
		if err != nil {
			return nil, fmt.Errorf("unable to delete existing file: %v", err)
		}
		println("Deleted existing file successfully")
	}
	f := &drive.File{
		Name:     fileName,
		MimeType: "application/octet-stream",
		Parents:  parents,
	}
	res, err := srv.Files.Create(f).Media(file).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %v", err)
	}
	return res, nil
}

func DeleteFile(srv *drive.Service, fileId string) error {
	err := srv.Files.Delete(fileId).Do()
	if err != nil {
		return fmt.Errorf("unable to delete file: %v", err)
	}
	fmt.Println("Deleted file:", fileId)
	return nil
}

func findFolder(srv *drive.Service, name string, parentId string) (string, error) {
	query := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false and 'root' in parents", name)
	if parentId != "" {
		query += fmt.Sprintf(" and '%s' in parents", parentId)
	}

	res, err := srv.Files.List().
		Q(query).
		Fields("files(id, name)").
		Do()
	if err != nil {
		return "", err
	}
	if len(res.Files) > 0 {
		return res.Files[0].Id, nil // return first match
	}
	return "", nil // not found
}

func FindFileByName(srv *drive.Service, name string) (*drive.File, error) {
	res, err := srv.Files.List().
		Q(fmt.Sprintf("name = '%s' and trashed = false", name)).
		Fields("files(id, name, mimeType)").
		Do()
	if err != nil {
		return nil, err
	}
	if len(res.Files) == 0 {
		return nil, fmt.Errorf("file '%s' not found", name)
	}
	return res.Files[0], nil
}

func DownloadFile(ctx context.Context, srv *drive.Service, fileID, localPath string) error {
	resp, err := srv.Files.Get(fileID).Download()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func GetLatestUploadTime(srv *drive.Service, worldName string) (string, error) {
	query := fmt.Sprintf("name = '%s.zip' and 'root' in parents and trashed = false", worldName)
	res, err := srv.Files.List().
		Q(query).
		Fields("files(id, name, modifiedTime)").
		OrderBy("modifiedTime desc").
		PageSize(1).
		Do()
	if err != nil {
		return "", err
	}
	if len(res.Files) == 0 {
		return "", fmt.Errorf("file '%s' not found", worldName)
	}
	return res.Files[0].ModifiedTime, nil
}

func UploadFolder(srv *drive.Service, filePath string, parentId string) (string, error) {
	// make the name = the last part of the path
	name := filePath
	if filePath != "" {
		parts := strings.Split(filePath, "/")
		name = parts[len(parts)-1]
	}
	existingId, err := findFolder(srv, name, parentId)
	if err != nil {
		return "", err
	}
	if existingId != "" {
		println("Folder already exists:", existingId)
		println("Deleting Existing Folder...")
		// folder exists delete it so we can replace it
		err = DeleteFile(srv, existingId)
		if err != nil {
			return "", fmt.Errorf("unable to delete existing folder: %v", err)
		}
		println("Deleted existing folder successfully")
	}
	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}
	if parentId != "" {
		folder.Parents = []string{parentId}
	}
	created, err := srv.Files.Create(folder).Fields("id").Do()
	if err != nil {
		return "", fmt.Errorf("unable to create folder: %v", err)
	}
	fmt.Println("Created folder:", name)
	// for every file in the folder upload it to this created folder
	entries, err := os.ReadDir(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read folder: %v", err)
	}
	for _, entry := range entries {
		println("Entry:", entry.Name())
		if entry.IsDir() {
			println("Recurring inside folder:", entry.Name())
			subFolderPath := filePath + "/" + entry.Name()
			subFolderId, err := UploadFolder(srv, subFolderPath, created.Id)
			if err != nil {
				return "", fmt.Errorf("unable to upload subfolder: %v", err)
			}
			println("Subfolder ID:", subFolderId)
			continue
		} else {
			filePath := filePath + "/" + entry.Name()
			file, err := os.Open(filePath)
			if err != nil {
				return "", fmt.Errorf("unable to open file: %v", err)
			}
			defer file.Close()
			ctx := context.Background()
			createdFile, err := UploadFile(ctx, srv, file, created.Id)
			if err != nil {
				return "", fmt.Errorf("unable to upload file: %v", err)
			}
			fmt.Println("Uploaded file:", createdFile.Name)
		}
	}
	return created.Id, nil
}

func InitDrive() (context.Context, *drive.Service, error) {
	// Create Drive service
	ctx := context.Background()
	b := credentialsJSON
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	client := getClient(config)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create Drive client: %w", err)
	}

	return ctx, srv, nil
}

func Authenticate() (string, error) {
	b := credentialsJSON
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getURL(config)

	return client, nil
}
