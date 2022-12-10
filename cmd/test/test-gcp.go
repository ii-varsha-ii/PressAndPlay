package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

const (
	GOOGLE_ACCESS_ID = "cloud-storage-object-sa@final-dsc-project.iam.gserviceaccount.com"
	PRIVATE_KEY      = "resources/private-key.pem"
	BUCKET_NAME      = "pressandplay"
)

func createBucket(client *storage.Client, bucketName string) (*storage.BucketHandle, error) {
	bkt := client.Bucket(bucketName)
	if err := bkt.Create(context.TODO(), "final-dsc-project", nil); err != nil {
		return nil, err
	}
	return bkt, nil
}

func deleteBucket(bkt *storage.BucketHandle) error {
	return bkt.Delete(context.TODO())
}

func uploadFile(googleAccessId string, privateKey []byte, bucket string, object string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("/Users/adarsh/development/src/github.com/adarshsrinivasan/PressAndPlay/resources/credentials.json"))
	if err != nil {
		log.Error(err)
	}
	defer client.Close()

	//bkt, err := createBucket(client, bucket)
	//defer deleteBucket(bkt)
	//if err != nil {
	//	log.Error(err)
	//	return err
	//}

	f, err := os.Open("resources/image.jpeg")
	if err != nil {
		log.Error(err)
		return err
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Minute*50)
	defer cancel()

	o := client.Bucket(bucket).Object(object)

	o = o.If(storage.Conditions{DoesNotExist: true})
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		log.Error(err)
	}
	if err := wc.Close(); err != nil {
		log.Error(err)
	}
	log.Info("Blob uploaded")
	opts := &storage.SignedURLOptions{
		GoogleAccessID: googleAccessId,
		PrivateKey:     []byte(privateKey),
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		Expires:        time.Now().Add(24 * time.Hour),
	}

	url, err := storage.SignedURL(bucket, object, opts)
	if err != nil {
		log.Error(err)
	}
	log.Info("Signed URL of the object: %s", url)
	return nil
}

func uploadImage(privateKey []byte, bucket, googleAccessId, key string) {
	signedURL := "https://storage.googleapis.com/pressandplay/d12e0fa6-7f57-45c1-bdc3-5f995a8dc92c?Expires=1670579440&GoogleAccessId=cloud-storage-object-sa%40final-dsc-project.iam.gserviceaccount.com&Signature=OErTesb2G55UE8I1BoxplgDemRYNLafvxYyUFXu73bzGaH2YtY8p%2Bb5AyzGNtbKf21aBQpHtZr8Z6BntpJlzDMXaruo8u%2FoIToJNUWlG3B6cLVed%2Bap5MlswLKUwX3stdx9F0emjm1asnXPtAzxwmNduybP6TfbRbrMRqi8eH72ZvJpw6FwWUzVNWMEBIIIZaGTOdmdPMeXMapnpSqnTlUGwTxex0N%2F67gqmr9P7LVJT5Hc0TnV%2FJcWY61KgzwrVkjpyUke7JrRr4G4Tbhr%2F7vUtW3RlTtYf0njgkA%2BQBdcpI2YJPeDLEpG2mJeUtA6I0DHkmyD8dQi3%2BOMAYo7mlQ%3D%3D"
	b, err := ioutil.ReadFile("resources/image.jpeg")
	if err != nil {
		log.Fatal(err)
	}
	// Generates *http.Request to request with PUT method to the Signed URL.
	req, err := http.NewRequest("PUT", signedURL, bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "image/png")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}

func genSignedURL(privateKey []byte, bucket, googleAccessId, key string) (string, error) {

	return storage.SignedURL(bucket, key, &storage.SignedURLOptions{
		GoogleAccessID: googleAccessId,
		Method:         "PUT",
		Expires:        time.Now().Add(15 * time.Minute),
		ContentType:    "image/png",
		PrivateKey:     privateKey,
	})
}

func main() {
	//privateKey, _ := ioutil.ReadFile(PRIVATE_KEY)
	//
	////objectName := "image2.jpeg"
	//key := uuid.New().String()
	//uploadImage(privateKey, BUCKET_NAME, GOOGLE_ACCESS_ID, key)
	newLayout := "1504"
	t1, err := time.Parse(newLayout, "1330")
	if err != nil {
		log.Fatal(err)
	}
	t2, err := time.Parse(newLayout, time.Now().Format(newLayout))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(t1.Sub(t2).Hours())
}
