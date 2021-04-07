package dbfuncs

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// drive types
const (
	DRIVE_IMAGE_TYPE = "image"
	DRIVE_AUDIO_TYPE = "audio"
	DRIVE_FILE_TYPE  = "file"
	DRIVE_VIDEO_TYPE = "video"
)

var fileService = &drive.FilesService{}
var DRIVE_IDS = map[string]string{
	"audio": "13pQa4gOuHZCc6f32evJdKdS16k1jSop-",
	"file":  "1xh4U-jxiZNbsurgH74eKFsn4R_eWujBB",
	"image": "1MnrTIUvci7hjL2-j9MlT_bUwQLbya6MT",
	"video": "18uR9YPRqGcpjYsSjMyCX-_fYHlurkLSQ",
	"db":    "1hD1X-98u7iTUp42V_FXwJF68EfWRfobY",
}

func initGoogleDriveService(eLogger *log.Logger) *drive.Service {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "drive-sa.json")
	service, e := drive.NewService(context.Background(), option.WithScopes(drive.DriveScope))
	if e != nil {
		eLogger.Fatal("Can not get service: " + e.Error())
	}
	return service
}

func initGoogleDriveFileService(eLogger *log.Logger) {
	service := initGoogleDriveService(eLogger)
	fileService = drive.NewFilesService(service)
}

func getFileByID(fileID string) ([]byte, error) {
	resp, e := fileService.Get(fileID).Download()
	if e != nil {
		return nil, errors.New("Can not get file")
	}

	fileData, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, errors.New("Can not read file")
	}
	return fileData, nil
}

func getDBFileFromGoogleDrive(eLogger *log.Logger) {
	fileData, e := getFileByID(DRIVE_IDS["db"])
	if e != nil {
		eLogger.Fatal(e)
	}

	e = ioutil.WriteFile("db/wnet.db", fileData, 0666)
	if e != nil {
		eLogger.Fatal("Can not create db file: " + e.Error())
	}
}

// UploadDB upload actual db to google drive(disk)
func UploadDB() error {
	file, e := os.Open("db/wnet.db")
	if e != nil {
		return e
	}

	_, e = fileService.Update(DRIVE_IDS["db"], &drive.File{}).Media(file).Do()
	return e
}

// CreateDriveFile get filetype with content and create google drive file in appropriate folder
func CreateDriveFile(typeFile, name string, file io.Reader) (*drive.File, error) {
	driveFile := &drive.File{
		Parents: []string{DRIVE_IDS[typeFile]},
		Name:    name,
	}
	return fileService.Create(driveFile).Media(file).Do()
}

// GetFileFromDrive get file from Google drive by id
func GetFileFromDrive(fileID string) ([]byte, error) {
	return getFileByID(fileID)
}
