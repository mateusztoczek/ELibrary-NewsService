// @title News API Documentation
// @version 1.0
// @description This is a sample API documentation for a news service.
// @host localhost:8080
// @schemes http

// @SecurityDefinitions
// Bearer:
//   type: apiKey
//   in: header
//   name: Authorization

// @Security
// - Bearer: []
package main

import (
	"log"
	"net/http"
	"news/config"
	"news/database"
	"news/handlers"

	_ "news/docs"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	err := RunServer()
	if err != nil {
		log.Fatal(err)
	}
}

func RunServer() error {

	config, err := config.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get config from file")
	}
	db, err := database.ConnectDB(config)
	if err != nil {
		return errors.Wrap(err, "failed to connect to the database")
	}
	defer db.Close()

	err = database.CreateNewsTable(db, config)
	if err != nil {
		return errors.Wrap(err, "failed to create News table")
	}

	router := mux.NewRouter()

	// Endpointy
	router.HandleFunc("/api/News", handlers.GetAllNews(db, config.SchemaName, config.TableName)).Methods("GET")
	router.HandleFunc("/api/News/{id}", handlers.GetNewsByID(db, config.SchemaName, config.TableName)).Methods("GET")
	router.HandleFunc("/api/News", handlers.CreateNews(db, config.SchemaName, config.TableName)).Methods("POST")
	router.HandleFunc("/api/News/{id}", handlers.UpdateNews(db, config.SchemaName, config.TableName)).Methods("PUT")
	router.HandleFunc("/api/News/{id}", handlers.DeleteNews(db, config.SchemaName, config.TableName)).Methods("DELETE")

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	log.Println("Serwer NewsService został uruchomiony na porcie 8080")
	return http.ListenAndServe(":8080", router)
}
