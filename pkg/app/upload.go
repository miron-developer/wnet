package app

import (
	"errors"
	"net/http"
	"wnet/pkg/orm"
)

func uploadFile(fileFormKey, fileType string, r *http.Request) (string, string, error) {
	file, fh, e := r.FormFile(fileFormKey)
	defer file.Close()
	if e != nil || fh == nil {
		return "", "", errors.New("file did not found")
	}

	if fh.Size/1024/1024 > 100 {
		return "", "", errors.New("this size is greater than 100mb")
	}

	link := "/assets/"
	driveFile, e := orm.CreateDriveFile(fileType, StringWithCharset(8)+fh.Filename, file)
	return link + driveFile.Id, fh.Filename, e
}
