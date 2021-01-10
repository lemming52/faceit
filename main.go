package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"faceit/service/dao"
	"faceit/service/handlers"
	"faceit/service/publisher"
)

const (
	// Service is the name of the service
	Service = "faceit-users"

	// Version is the current version of the service
	Version = "0.0.1"

	// UsersURI is the address for the add and filter operations
	UsersURI = "/users"

	// SingleUserURI is the address for any operation on a given user ID
	SingleUserURI = "/users/{id}"

	// HealthCheckURI is the uri for the basic status endpoint
	HealthCheckURI = "/healthcheck"

	// DocsURI is the endpoint for the prerendered documentation
	DocsURI = "/docs"

	// Host is the hardcoded local host for the service
	// Note that changing the port will require alterations to the dockerfile and docker-compose
	Host = "0.0.0.0:3000"
)

func main() {
	log.SetFormatter(&logrus.JSONFormatter{})

	log.Info("start server")
	r := mux.NewRouter()

	db := getDatabase()
	msg := getPublisher()

	h := handlers.NewHandler(db, msg)
	r.HandleFunc(DocsURI, handlers.GetDocHandler(handlers.DocPath)).Methods(http.MethodGet)
	r.HandleFunc(HealthCheckURI, handlers.GetHealthCheckHandler(Service, Version))

	r.HandleFunc(SingleUserURI, handlers.ToHandlerFunc(h.RemoveUser)).Methods(http.MethodDelete)
	r.HandleFunc(SingleUserURI, handlers.ToHandlerFunc(h.UpdateUser)).Methods(http.MethodPut)
	r.HandleFunc(SingleUserURI, handlers.ToHandlerFunc(h.GetUser)).Methods(http.MethodGet)

	r.HandleFunc(UsersURI, handlers.ToHandlerFunc(h.AddUser)).Methods(http.MethodPost)
	r.HandleFunc(UsersURI, handlers.ToHandlerFunc(h.FilterUsers)).Methods(http.MethodGet)

	server := &http.Server{
		Handler: r,
		Addr:    Host,
	}
	log.Fatal(server.ListenAndServe())
}

func getDatabase() *dao.DynamoClient {
	return dao.NewDynamoClient()
}

func getPublisher() *publisher.SNSClient {
	return publisher.NewSNSClient()
}
