package main

import "net/http"

type writer interface {
	Read(string) ([]byte, error)
	Write(*http.Request, string) ([]byte, error)
	Delete(string) ([]byte, error)
}
