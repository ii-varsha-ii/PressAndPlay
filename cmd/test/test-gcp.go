package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
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
	signedURL := "https://storage.googleapis.com/pressandplay/f9ad34be-f459-4133-a4f0-d9545fce17e0?Expires=1670776204&GoogleAccessId=cloud-storage-object-sa%40final-dsc-project.iam.gserviceaccount.com&Signature=IDefwFZtG3txdfLlqQlN%2BrFyHtXa9QwvJzflh6Zih4mu23ia8CzECa3hDvZ82zvfl%2B%2Fftzpi37eC6BbtXqgwfu8tUk5TwdLeeT5woCifh6ZkzxnD%2Bz%2BRbeI8BCtmRsUbNiaNw1y0NxeM39Nb6XVoT%2BLn0G4Z6ktoTKGloIeC%2FV6%2BF5FGNrOWMI9B54CUqMDybCBxQeiBXfiHcP3d7IvTFEFOyJD5wVaMMyyJaUnDZJBYwDFCfhYhdTfb8MuHlTz8bZaeQWaGBv8We6pJcegIz%2BLu2tbpxiUW9y62Pd3QdXTrpl2UGBLvryzDdp6Z0yrEq6cIVdsDCxahsnC6p14DIA%3D%3D"
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

func main() {
	privateKey, _ := ioutil.ReadFile(PRIVATE_KEY)

	//objectName := "image2.jpeg"
	key := uuid.New().String()
	uploadImage(privateKey, BUCKET_NAME, GOOGLE_ACCESS_ID, key)
}
