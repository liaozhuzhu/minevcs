package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

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

func UploadFile(ctx context.Context, srv *drive.Service, file *os.File) (*drive.File, error) {
	f := &drive.File{
		Name:     file.Name(),
		MimeType: "application/octet-stream",
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

func uploadFolder(srv *drive.Service, name string, parentId string) (string, error) {
	existingId, err := findFolder(srv, name, parentId)
	if err != nil {
		return "", err
	}
	if existingId != "" {
		println("Folder already exists:", existingId)
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
	return created.Id, nil
}

func uploadDirectoryToDrive(srv *drive.Service, localPath string, parentDriveID string) error {
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the base directory (already created as parentDriveID)
		if path == localPath {
			return nil
		}

		relPath, _ := filepath.Rel(localPath, path)

		if info.IsDir() {
			// Create Drive folder
			folder := &drive.File{
				Name:     info.Name(),
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{parentDriveID},
			}
			created, err := srv.Files.Create(folder).Fields("id").Do()
			if err != nil {
				return fmt.Errorf("could not create folder %s: %w", relPath, err)
			}
			// Recurse into subdirectory
			return uploadDirectoryToDrive(srv, path, created.Id)
		} else {
			// Upload file
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			driveFile := &drive.File{
				Name:    info.Name(),
				Parents: []string{parentDriveID},
			}
			_, err = srv.Files.Create(driveFile).
				Media(f).
				Fields("id").
				Do()
			if err != nil {
				return fmt.Errorf("upload failed for %s: %w", relPath, err)
			}
			fmt.Println("Uploaded:", relPath)
		}
		return nil
	})
}

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	// shortFile, err := os.Open("test.txt")
	// outFile, err := UploadFile(ctx, srv, shortFile)
	// if err != nil {
	// 	log.Fatalf("Unable to upload file: %v", err)
	// }
	// fmt.Printf("File '%s' uploaded with ID: %s\n", outFile.Name, outFile.Id)

	// r, err := srv.Files.List().PageSize(10).
	// 	Fields("nextPageToken, files(id, name)").Do()
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve files: %v", err)
	// }
	// fmt.Println("Files:")
	// if len(r.Files) == 0 {
	// 	fmt.Println("No files found.")
	// } else {
	// 	for _, i := range r.Files {
	// 		fmt.Printf("%s (%s)\n", i.Name, i.Id)
	// 	}
	// }

	// Define the parent folder in Drive
	folderId, err := uploadFolder(srv, "test", "")
	if err != nil {
		log.Fatalf("Unable to create folder: %v", err)
	}
	println("Upload folder with ID:", folderId)
}
