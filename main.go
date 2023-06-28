package main

import (
	"log"
	"net/http"
	"news/config"
	"news/database"
	"news/routes"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/gorilla/handlers"
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
	methods := handlers.AllowedMethods([]string{"OPTIONS", "DELETE", "GET", "HEAD", "POST", "PUT"})
	origins := handlers.AllowedOrigins([]string{"*"})
	credentials := handlers.AllowCredentials()

	// Endpointy
	router.HandleFunc("/api/News", routes.GetAllNews(db, config.SchemaName, config.TableName)).Methods("GET")
	router.HandleFunc("/api/News/{id}", routes.GetNewsByID(db, config.SchemaName, config.TableName)).Methods("GET")
	router.HandleFunc("/api/News", routes.CreateNews(db, config.SchemaName, config.TableName)).Methods("POST")
	router.HandleFunc("/api/News/{id}", routes.UpdateNews(db, config.SchemaName, config.TableName)).Methods("PUT")
	router.HandleFunc("/api/News/{id}", routes.DeleteNews(db, config.SchemaName, config.TableName)).Methods("DELETE")

	log.Println("Serwer NewsService został uruchomiony na porcie 8080")
	return http.ListenAndServe(":8080", handlers.CORS(credentials, methods, origins)(router))
}
