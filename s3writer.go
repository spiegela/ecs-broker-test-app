package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Writer struct {
	bucket string
	svc    *s3.S3
}

func NewS3Writer(bucket string, svc *s3.S3) *s3Writer {
	return &s3Writer{
		bucket: bucket,
		svc:    svc,
	}
}

func (w s3Writer) Delete(key string) ([]byte, error) {
	resp, s3Err := w.svc.DeleteObjectWithContext(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(w.bucket),
		Key:    aws.String(key),
	})
	if s3Err != nil {
		return nil, s3Err
	}
	return []byte(resp.String()), nil
}

func (w s3Writer) Read(key string) ([]byte, error) {
	s3Resp, s3Err := w.svc.GetObjectWithContext(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(w.bucket),
		Key:    aws.String(key),
	})
	if s3Err != nil {
		return nil, s3Err
	}
	content, readErr := ioutil.ReadAll(s3Resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	return content, nil
}

func (w s3Writer) Write(r *http.Request, key string) ([]byte, error) {
	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		return nil, readErr
	}
	resp, s3Err := w.svc.PutObjectWithContext(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(w.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	})
	if s3Err != nil {
		return nil, s3Err
	}
	return []byte(resp.String()), nil
}
