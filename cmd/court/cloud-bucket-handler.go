package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"io/ioutil"
	"time"

	"cloud.google.com/go/storage"
	"github.com/adarshsrinivasan/PressAndPlay/libraries/common"
)

const (
	GCP_PRIVATE_KEY_PATH_ENV   = "GCP_PRIVATE_KEY_PATH"
	GCP_BUCKET_NAME_ENV        = "GCP_BUCKET_NAME"
	GCP_ACCESS_ID_ENV          = "GCP_ACCESS_ID"
	GCP_IMAGE_URL_TEMPLATE_ENV = "GCP_IMAGE_URL_TEMPLATE"
	GCP_CREDENTIALS_PATH_ENV   = "GCP_CREDENTIALS_PATH"
)

func deleteObjectFromCloud(key string) error {
	credentialsPath := common.GetEnv(GCP_CREDENTIALS_PATH_ENV, "resources/credentials.json")
	bucketName := common.GetEnv(GCP_BUCKET_NAME_ENV, "pressandplay")
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Error(err)
	}
	defer client.Close()
	bkt := client.Bucket(bucketName)
	if err := bkt.Create(context.TODO(), "final-dsc-project", nil); err != nil {
		return err
	}
	return bkt.Object(key).Delete(context.TODO())
}

func genImageDownloadAndUploadUrl(key string) (string, string, error) {
	privateKeyPath := common.GetEnv(GCP_PRIVATE_KEY_PATH_ENV, "resources/private-key.pem")
	bucketName := common.GetEnv(GCP_BUCKET_NAME_ENV, "pressandplay")
	googleAccessId := common.GetEnv(GCP_ACCESS_ID_ENV, "cloud-storage-object-sa@final-dsc-project.iam.gserviceaccount.com")
	imageUrlTemplate := common.GetEnv(GCP_IMAGE_URL_TEMPLATE_ENV, "https://storage.cloud.google.com/pressandplay/%s?authuser=6")

	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return "", "", err
	}
	signedUrl, err := storage.SignedURL(bucketName, key, &storage.SignedURLOptions{
		GoogleAccessID: googleAccessId,
		Method:         "PUT",
		Expires:        time.Now().Add(15 * time.Minute),
		ContentType:    "image/png",
		PrivateKey:     privateKey,
	})
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf(imageUrlTemplate, key), signedUrl, nil
}
