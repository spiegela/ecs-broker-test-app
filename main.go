package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const vcapServicesVar = "VCAP_SERVICES"

const fullControlACL = `<?xml version="1.0" encoding="UTF-8"?>
<AccessControlPolicy xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Owner>
    <ID>{access_key}</ID>
    <DisplayName>owner-display-name</DisplayName>
  </Owner>
  <AccessControlList>
    <Grant>
      <Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
               xsi:type="Canonical User">
        <ID>{access_key}</ID>
        <DisplayName>{access_key}</DisplayName>
      </Grantee>
      <Permission>FULL_CONTROL</Permission>
    </Grant>
  </AccessControlList>
</AccessControlPolicy>`

var (
	s3Client *s3.S3
	bucket   string
)

func main() {
	svc, bucketName, err := generateS3Client()
	if err != nil {
		log.Fatal(err)
	}
	s3Client = svc
	bucket = bucketName
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	var (
		key  = r.URL.Path[1:]
		err  error
		resp []byte
	)
	asFile := r.URL.Query().Get("file")
	var wr writer
	if asFile != "" {
		wr = NewFileWriter(bucket)
	} else {
		wr = NewS3Writer(bucket, s3Client)
	}
	switch r.Method {
	case http.MethodPut, http.MethodPost:
		resp, err = wr.Write(r, key)
	case http.MethodGet:
		resp, err = wr.Read(key)
	case http.MethodDelete:
		resp, err = wr.Delete(key)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, writeRespErr := w.Write([]byte(err.Error()))
		if writeRespErr != nil {
			log.Fatal(writeRespErr)
		}
	}
	_, writeRespErr := w.Write(resp)
	if writeRespErr != nil {
		log.Fatal(writeRespErr)
	}
}

func generateS3Client() (*s3.S3, string, error) {
	vcapServices := os.Getenv(vcapServicesVar)
	services := bindingLookup{}
	err := json.Unmarshal([]byte(vcapServices), &services)
	if err != nil {
		return nil, "", err
	}

	bucketServices, bucketServiceDefined := services["ecs-bucket"]
	namespaceServices, namespaceServiceDefined := services["ecs-namespace"]
	fileBucketServices, fileBucketServiceDefined := services["ecs-file-bucket"]

	var serviceList []serviceBinding
	switch {
	case bucketServiceDefined:
		serviceList = bucketServices
	case namespaceServiceDefined:
		serviceList = namespaceServices
	case fileBucketServiceDefined:
		serviceList = fileBucketServices
	}
	if len(serviceList) == 0 {
		return nil, "", fmt.Errorf("no services defined for service definiton: %v", serviceList)
	}
	binding := serviceList[0]
	endpoint, sess, err := createSessionFromCredentials(binding.Credentials)
	if err != nil {
		return nil, "", err
	}
	forcePathStyle := true
	svc := s3.New(sess, &aws.Config{Endpoint: &endpoint, S3ForcePathStyle: &forcePathStyle})
	bucket = binding.Credentials.Bucket
	if bucket == "" {
		bucket = "test-bucket"
		acl, err := generateFullControlACL(binding.Credentials.AccessKey)
		if err != nil {
			return nil, "", err
		}
		_, err = svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: &bucket,
			ACL:    &acl,
		})
		if err != nil {
			return nil, "", err
		}
	}
	return svc, bucket, nil
}

func createSessionFromCredentials(serviceCredentials serviceCredentials) (string, *session.Session, error) {
	sess, sessErr := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(
			serviceCredentials.AccessKey,
			serviceCredentials.SecretKey,
			"TOKEN",
		),
		Endpoint: &serviceCredentials.Endpoint,
	})
	if sessErr != nil {
		return "", nil, sessErr
	}
	return serviceCredentials.Endpoint, sess, nil
}

func generateFullControlACL(accessKey string) (string, error) {
	var out bytes.Buffer
	tmpl, err := template.New("fullControlACL").Parse(fullControlACL)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&out, struct{ accessKey string }{
		accessKey: accessKey,
	})
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
