package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const vcapServicesVar = "VCAP_SERVICES"

var (
	bucket string
	svc    *s3.S3
	ctx    = context.Background()
)

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]

	switch r.Method {
	case http.MethodPut, http.MethodPost:
		body, readErr := ioutil.ReadAll(r.Body)
		if readErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, writeRespErr := w.Write([]byte(readErr.Error()))
			log.Fatal(writeRespErr)
		}
		resp, s3Err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   bytes.NewReader(body),
		})
		if s3Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, writeRespErr := w.Write([]byte(s3Err.Error()))
			if writeRespErr != nil{
				log.Fatal(writeRespErr)
			}
			return
		}
		_, writeRespErr := w.Write([]byte(resp.String()))
		if writeRespErr != nil {
			log.Fatal(writeRespErr)
		}
	case http.MethodGet:
		s3Resp, s3Err := svc.GetObjectWithContext(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if s3Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, writeRespErr := w.Write([]byte(s3Err.Error()))
			if writeRespErr != nil {
				log.Fatal(writeRespErr)
			}
			return
		}
		content, readErr := ioutil.ReadAll(s3Resp.Body)
		if readErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, writeRespErr := w.Write([]byte(readErr.Error()))
			if writeRespErr != nil {
				log.Fatal(writeRespErr)
			}
		}
		_, writeRespErr := w.Write(content)
		if writeRespErr != nil {
			log.Fatal(writeRespErr)
		}
	case http.MethodDelete:
		resp, s3Err := svc.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if s3Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, writeRespErr := w.Write([]byte(s3Err.Error()))
			log.Fatal(writeRespErr)
			return
		}
		_, writeRespErr := w.Write([]byte(resp.String()))
		if writeRespErr != nil {
			log.Fatal(writeRespErr)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	vcapServices := os.Getenv(vcapServicesVar)
	var services map[string][]map[string]interface{}
	err := json.Unmarshal([]byte(vcapServices), &services)
	if err != nil {
		log.Fatal(err)
	}
	vcapCredentials := services["ecs-bucket"][0]["credentials"].(map[string]interface{})
	if len(services["ecs-bucket"]) == 0 {
		log.Fatalf("No bucket bound: %v", vcapServices)
	}
	endpoint := vcapCredentials["endpoint"].(string)
	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(
			vcapCredentials["accessKey"].(string),
			vcapCredentials["secretKey"].(string),
			"TOKEN",
		),
		Endpoint: &endpoint,
	})
	if sessErr != nil {
		log.Fatal(sessErr)
	}
	forcePathStyle := true
	svc = s3.New(sess, &aws.Config{Endpoint: &endpoint, S3ForcePathStyle: &forcePathStyle})
	bucket = vcapCredentials["bucket"].(string)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
