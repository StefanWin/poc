package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Api struct {
	Router *mux.Router
	Server *http.Server
	// chan<- means write-only
	RequestChannel chan<- *ConversionRequest
}

func NewApi(requestChannel chan<- *ConversionRequest) (*Api, error) {
	router := mux.NewRouter()
	server := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:3000", // TODO : extract to config
		WriteTimeout: time.Second * 15, // TODO : extract to config
		ReadTimeout:  time.Second * 15, // TODO : extract to config
	}
	api := &Api{
		Router:         router,
		Server:         server,
		RequestChannel: requestChannel,
	}
	api.configureRoutes()
	return api, nil
}

// TODO : move to new file 'middleware.go'
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API]:: %s\n", r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func (api *Api) configureRoutes() {
	api.Router.HandleFunc("/conversion/v2", api.ConvertFileHandler).Methods("POST")
	api.Router.HandleFunc("/conversion/", api.ConversionQueueStatusHandler).Methods("GET")
	api.Router.HandleFunc("/conversion/{conversionId}", api.GetConvertedFileHandler).Methods("GET")
	api.Router.HandleFunc("/conversion/{conversionId}/download", api.GetConvertedFileDownloadHandler).Methods("GET")
	api.Router.Use(loggingMiddleware)
}

func (api *Api) Listen() error {
	log.Println("[API]:: listening on 127.0.0.1:3000")
	return api.Server.ListenAndServe()
}

func (api *Api) Shutdown() {
	// TODO : Remove dangling files from input
	log.Println("[API]:: shutting down...")
	api.Server.Shutdown(context.Background())
}
