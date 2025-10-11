package utils

import (
	"context"
	"errors"
	"mime/multipart"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadFileToCloud(file *multipart.FileHeader, folder, publicID string) (string, error) {
	if file == nil {
		return "", errors.New("no file provided")
	}

	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	uploadResult, err := cld.Upload.Upload(context.Background(), src, uploader.UploadParams{
		Folder:       "enerzyflow/" + folder, 
		ResourceType: "raw",                 
		PublicID:     publicID,
	})
	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}

func NowInIST() time.Time {
	ist := time.FixedZone("IST", 5*60*60+30*60)
	return time.Now().In(ist)
}