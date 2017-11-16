package main

import (
	"errors"
	"github.com/disintegration/imaging"
	"github.com/satori/go.uuid"
	"os"
	"strconv"
	"strings"
	"github.com/6thplaneta/hermes"
	"time"
)

var typeToExt = map[string]string{
	"text/css; charset=utf-8":  ".css",
	"image/gif":                ".gif",
	"text/html; charset=utf-8": ".html",
	"image/jpeg":               ".jpg",
	"application/x-javascript": ".js",
	"application/pdf":          ".pdf",
	"image/png":                ".png",
	"image/svg+xml":            ".svg",
	"text/xml; charset=utf-8":  ".xml",
}

func GetAllUrls(contentType string) (string, string, string, string, error) {
	uniquefid := uuid.NewV4().String()

	year, month, _ := time.Now().Date()
	var path_, url string
	path_ = "./upload/"
	url = "/upload/"
	savePath := application.Conf.GetString("public.upload")
	if savePath != "" {
		path_ = savePath
	}
	parts := strings.Split(contentType, "/")
	path_ += parts[0] + "/"
	url += parts[0] + "/"
	path_ += strconv.Itoa(year) + "/" + strconv.Itoa(int(month)) + "/"
	url += strconv.Itoa(year) + "/" + strconv.Itoa(int(month)) + "/"
	path100 := path_ + "100x100/"
	path380 := path_ + "380x380/"
	path640 := path_ + "640x640/"
	url += "640x640/"
	err := os.MkdirAll(path100, 0777)
	if err != nil {
		return "", "", "", "", err
	}
	err = os.MkdirAll(path380, 0777)
	if err != nil {
		return "", "", "", "", err
	}
	err = os.MkdirAll(path640, 0777)
	if err != nil {
		return "", "", "", "", err
	}
	ext := typeToExt[contentType]

	url100 := path100 + uniquefid + ext
	url380 := path380 + uniquefid + ext
	url640 := path640 + uniquefid + ext
	url = url + uniquefid + ext
	return url100, url380, url640, url, nil

}

type UploadRecord struct {
	Id  int
	Url string
}

func uploadImageMW(url, content_type string) (interface{}, error) {

	if content_type != "image/jpeg" {
		return "", errors.New("image should be jpeg")
	}

	url100, url380, url640, urlOut, err := GetAllUrls(content_type)

	if err != nil {
		return "", err
	}
	mainImage, err := imaging.Open(url)

	Image640 := imaging.Resize(mainImage, 640, 640, imaging.Lanczos)
	err = imaging.Save(Image640, url640)
	if err != nil {
		return "", err
	}

	Image380 := imaging.Resize(mainImage, 380, 380, imaging.Lanczos)
	err = imaging.Save(Image380, url380)
	if err != nil {
		return "", err
	}
	Image100 := imaging.Resize(mainImage, 100, 100, imaging.Lanczos)
	err = imaging.Save(Image100, url100)
	if err != nil {
		return "", err
	}

	//save address in db
	f := &File{}
	f.File_Path = urlOut
	f.Creation_Date = time.Now()
	uprec := UploadRecord{}
	obj, err := FileColl.Create(hermes.SystemToken, nil, f)
	if err != nil {
		return uprec, err
	}
	fcreated := obj.(*File)
	uprec.Id = fcreated.Id
	uprec.Url = urlOut
	err = os.Remove(url)
	if err != nil {
		application.Logger.Error("Remove Original Image:" + err.Error())
	}
	return uprec, nil
}
