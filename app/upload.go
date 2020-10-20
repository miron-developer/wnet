package app

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

// StringWithCharset - random string
func StringWithCharset(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func createFile(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
}

func uploadPhoto(file *multipart.File, filename, ext string) error {
	f, e := createFile(filename)
	if e != nil {
		return e
	}
	defer f.Close()
	var img image.Image

	if ext == "image/jpg" || ext == "image/jpeg" {
		img, e = jpeg.Decode(*file)
		e = jpeg.Encode(f, img, nil)
	} else if ext == "image/png" {
		img, e = png.Decode(*file)
		e = png.Encode(f, img)
	} else if ext != "image/gif" {
		img, e = gif.Decode(*file)
		e = gif.Encode(f, img, nil)
	} else {
		os.Remove(filename)
		return errors.New("dont support this type of photo")
	}
	return e
}

func uploadFile(fileGet, filePlace, fileType string, r *http.Request) (string, error) {
	file, fh, e := r.FormFile(fileGet)
	defer file.Close()
	if e != nil {
		return "", e
	}

	if fh == nil {
		return "", errors.New("file not founded")
	}

	if strings.Contains(r.PostFormValue("oldphoto"), fh.Filename) {
		return "", errors.New("this file exist")
	}

	if fh.Size/1024/1024 > 20 {
		return "", errors.New("this size is greater than 20mb")
	}

	fileExt := fh.Header.Get("Content-Type")
	filePreName := StringWithCharset(8)
	link := "static/" + fileType + "/" + filePlace + "/" + filePreName + fh.Filename
	if fileType == "img" {
		e = uploadPhoto(&file, link, fileExt)
	}
	return link, e
}
