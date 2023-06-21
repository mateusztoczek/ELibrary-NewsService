package main

import (
	"log"
	"net/http"
	"news/database"
	"news/handlers"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func main() {
	err := RunServer()
	if err != nil {
		log.Fatal(err)
	}
}

func RunServer() error {

	db, err := database.ConnectDB()
	if err != nil {
		return errors.Wrap(err, "failed to connect to the database")
	}
	defer db.Close()

	err = database.CreateNewsTable(db)
	if err != nil {
		return errors.Wrap(err, "failed to create News table")
	}

	router := mux.NewRouter()

	// Endpointy
	router.HandleFunc("/api/News", handlers.GetAllNews(db)).Methods("GET")
	router.HandleFunc("/api/News/{id}", handlers.GetNewsByID(db)).Methods("GET")
	router.HandleFunc("/api/News", handlers.CreateNews(db)).Methods("POST")
	router.HandleFunc("/api/News/{id}", handlers.UpdateNews(db)).Methods("PUT")
	router.HandleFunc("/api/News/{id}", handlers.DeleteNews(db)).Methods("DELETE")

	log.Println("Serwer NewsService został uruchomiony na porcie 8080")
	return http.ListenAndServe(":8080", router)
}
