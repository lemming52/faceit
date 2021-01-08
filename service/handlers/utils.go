package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const DocPath = "./docs/index.html"

type ErrorResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

type EndpointFunc func(r *http.Request) (int, interface{}, error)

func ToHandlerFunc(e EndpointFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code, payload, err := e(r)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(code)
			response := errorToResponse(code, err)
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				log.Fatal("error encoding error")
			}
			return
		}
		w.WriteHeader(code)
		if payload != nil {
			err = json.NewEncoder(w).Encode(payload)
			if err != nil {
				log.Fatal(fmt.Sprintf("error encoding payload: %v", err))
			}
		}
	}
}

func errorToResponse(code int, err error) *ErrorResponse {
	return &ErrorResponse{
		Code:        code,
		Description: err.Error(),
	}
}

type HealthCheck struct {
	service string `json:"service"`
	version string `json:"version"`
}

func GetHealthCheckHandler(service, version string) func(w http.ResponseWriter, r *http.Request) {
	return ToHandlerFunc(func(r *http.Request) (int, interface{}, error) {
		return http.StatusOK, &HealthCheck{
			service: service,
			version: version,
		}, nil
	})
}

func GetDocHandler(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
