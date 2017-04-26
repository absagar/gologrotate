package logrotate

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
	"os"
	"path/filepath"
	"time"
)

func DeleteBackup(origFile, backupFile string) error {
	return os.Remove(backupFile)
}

func awsBackup(file string) error {
	//TODO
	return nil
}

type GceConfig struct {
	CredentialsFile string
	Scope           string
	Bucket          string
	Location        string
}

func (g *GceConfig) Backup(origFile, backupFile string) error {
	debugln("Backing file :", backupFile)

	jsonFile, _ := filepath.Abs(g.CredentialsFile)
	if jsonFile != "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", jsonFile)
	}

	client, err := google.DefaultClient(context.Background(), g.Scope)
	if err != nil {
		return err
	}

	service, err := storage.New(client)
	if err != nil {
		return err
	}

	hostName, _ := os.Hostname()
	_, fileObj := filepath.Split(origFile)

	locationWithSlash := g.Location
	if locationWithSlash != "" && locationWithSlash[len(locationWithSlash)-1] != '/' {
		locationWithSlash = locationWithSlash + "/"
	}

	objectName := locationWithSlash + fileObj + "/" + hostName + time.Now().String()

	object := &storage.Object{Name: objectName}
	file, err := os.Open(backupFile)
	defer file.Close()
	if err != nil {
		return err
	}
	if _, err := service.Objects.Insert(g.Bucket, object).Media(file).Do(); err != nil {
		return err
	}

	os.Remove(backupFile)

	return nil
}