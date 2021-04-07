package app

import (
	"errors"
	"net/http"
	"wnet/app/dbfuncs"
)

func uploadFile(fileFormKey, fileType string, r *http.Request) (string, error) {
	file, fh, e := r.FormFile(fileFormKey)
	defer file.Close()
	if e != nil || fh == nil {
		return "", errors.New("file did not found")
	}

	if fh.Size/1024/1024 > 100 {
		return "", errors.New("this size is greater than 100mb")
	}

	link := "/assets/"
	driveFile, e := dbfuncs.CreateDriveFile(fileType, StringWithCharset(8)+fh.Filename, file)
	return link + driveFile.Id, e
}
