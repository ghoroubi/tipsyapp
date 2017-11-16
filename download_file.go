package main

import (
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"image"
	"io"
	"net/http"
	"os"
	"github.com/6thplaneta/hermes"
	"time"
)

func getImageDimension(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err

	}
	return image.Width, image.Height, nil
}

type HttpUrl struct {
	Url string `json:"url"`
}

func Download(c *gin.Context) {
	var h HttpUrl
	err := c.BindJSON(&h)
	if err != nil {
		hermes.HandleHttpError(c, hermes.ErrJsonFormat, application.Logger)
		return
	}
	url, err := DownloadAndSave(h.Url)
	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusCreated, url)
}

func DownloadAndSave(strUrl string) (UploadRecord, error) {
	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := check.Get(strUrl) // add a filter to check redirect
	if err != nil {
		return UploadRecord{}, err
	}
	defer resp.Body.Close()

	buff := make([]byte, 512) // why 512 bytes ? see http://golang.org/pkg/net/http/#DetectContentType
	_, err = resp.Body.Read(buff)

	contentType := http.DetectContentType(buff)

	url100, url380, url640, url, err := GetAllUrls(contentType)
	if err != nil {
		return UploadRecord{}, err
	}
	file, err := os.Create(url640)

	if err != nil {
		return UploadRecord{}, err
	}
	defer file.Close()

	resp, err = check.Get(strUrl) // add a filter to check redirect
	if err != nil {
		return UploadRecord{}, err
	}

	_, err = io.Copy(file, resp.Body)

	if err != nil {
		return UploadRecord{}, err
	}

	Imagemain, err := imaging.Open(url640)

	w, h, _ := getImageDimension(url640)
	var cropCenter *image.NRGBA
	if w != h {
		if w > h {
			cropCenter = imaging.CropCenter(Imagemain, h, h)
			err = imaging.Save(cropCenter, url640)
		} else {
			cropCenter = imaging.CropCenter(Imagemain, w, w)
			err = imaging.Save(cropCenter, url640)
		}
	}

	Imagemain, err = imaging.Open(url640)

	Image640 := imaging.Resize(Imagemain, 640, 640, imaging.Lanczos)
	err = imaging.Save(Image640, url640)
	if err != nil {
		return UploadRecord{}, err
	}

	Image380 := imaging.Resize(Imagemain, 380, 380, imaging.Lanczos)
	err = imaging.Save(Image380, url380)
	if err != nil {
		return UploadRecord{}, err
	}

	Image100 := imaging.Resize(Imagemain, 100, 100, imaging.Lanczos)
	err = imaging.Save(Image100, url100)
	if err != nil {
		return UploadRecord{}, err
	}

	//save address in db
	f := &File{}
	f.File_Path = url
	f.Creation_Date = time.Now()
	uprec := UploadRecord{}
	obj, err := FileColl.Create(hermes.SystemToken, nil, f)
	if err != nil {
		return uprec, err
	}
	fcreated := obj.(*File)
	uprec.Id = fcreated.Id
	uprec.Url = url

	return uprec, nil
}
