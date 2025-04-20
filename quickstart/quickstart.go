package quickstart

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// built off of quickstart example from google drive api w/ go (adjusted for UX)
func getClient(config *oauth2.Config) *http.Client {
	// local token file should already exist by now
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		println(err)
	}
	return config.Client(context.Background(), tok)
}

func getURL(config *oauth2.Config) string {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	_, err := tokenFromFile(tokFile)
	if err != nil {
		// tok = getTokenFromWeb(config)
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		return authURL
	}
	// return config.Client(context.Background(), tok)
	return ""
}

func HandleAuthCode(code string) *http.Client {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	tokFile := "token.json"
	saveToken(tokFile, tok)
	return config.Client(context.Background(), tok)
}

// // Request a token from the web, then returns the retrieved token.
// func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
// 	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
// 	fmt.Printf("Go to the following link in your browser then type the "+
// 		"authorization code: \n%v\n", authURL)

// 	var authCode string
// 	if _, err := fmt.Scan(&authCode); err != nil {
// 		log.Fatalf("Unable to read authorization code %v", err)
// 	}

// 	tok, err := config.Exchange(context.TODO(), authCode)
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve token from web %v", err)
// 	}
// 	return tok
// }

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
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
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// BEGIN GOOGLE DRIVE API
func UploadFile(ctx context.Context, srv *drive.Service, file *os.File, parentId string) (*drive.File, error) {
	fileName := strings.Split(file.Name(), "/")[len(strings.Split(file.Name(), "/"))-1]
	f := &drive.File{
		Name:     fileName,
		MimeType: "application/octet-stream",
		Parents:  []string{parentId},
	}
	res, err := srv.Files.Create(f).Media(file).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %v", err)
	}
	return res, nil
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
		err = srv.Files.Delete(existingId).Do()
		if err != nil {
			return "", fmt.Errorf("unable to delete existing folder: %v", err)
		}
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

func InitDrive() (*drive.Service, error) {
	// Create Drive service
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	client := getClient(config)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Drive client: %w", err)
	}

	return srv, nil
}

func Authenticate() (string, error) {
	// ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getURL(config)

	// _, err = drive.NewService(ctx, option.WithHTTPClient(client))
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve Drive client: %v", err)
	// }
	// println("New token created successfully")
	return client, nil
}
