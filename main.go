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

const Service = "faceit-users"
const Version = "0.0.1"
const UsersAPI = "/users"
const SingleUserAPI = "/users/{id}"

func main() {

	log.SetFormatter(&logrus.JSONFormatter{})
	log.Info("start server")
	r := mux.NewRouter()
	r.HandleFunc("/healthcheck", handlers.GetHealthCheckHandler(Service, Version))

	db := getDatabase()
	msg := getPublisher()

	h := handlers.NewHandler(db, msg)

	r.HandleFunc(UsersAPI, handlers.ToHandlerFunc(h.AddUser)).Methods(http.MethodPost)
	r.HandleFunc(SingleUserAPI, handlers.ToHandlerFunc(h.RemoveUser)).Methods(http.MethodDelete)
	r.HandleFunc(SingleUserAPI, handlers.ToHandlerFunc(h.UpdateUser)).Methods(http.MethodPut)
	r.HandleFunc(SingleUserAPI, handlers.ToHandlerFunc(h.GetUser)).Methods(http.MethodGet)
	r.HandleFunc(UsersAPI, handlers.ToHandlerFunc(h.FilterUsers)).Methods(http.MethodGet)
	r.HandleFunc("/docs", handlers.GetDocHandler(handlers.DocPath)).Methods(http.MethodGet)

	server := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:3000",
	}
	log.Fatal(server.ListenAndServe())
}

func getDatabase() *dao.DynamoClient {
	return dao.NewDynamoClient()
}

func getPublisher() *publisher.SNSClient {
	return publisher.NewSNSClient()
}
