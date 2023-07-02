package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"news/config"
	"testing"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var serverInstance *http.Server

// Test creating local server with db connection
func TestRunServer(t *testing.T) {

	//wczytaj konfiguracje
	configData, err := ioutil.ReadFile("../config/testConfig.json")
	if err != nil {
		t.Fatal("failed to read config file:", err)
	}

	var testConfig config.Config
	err = json.Unmarshal(configData, &testConfig)
	if err != nil {
		t.Fatal("failed to parse config file:", err)
	}

	//połącz z bazą danych Postgres
	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		testConfig.Host, testConfig.Port, testConfig.User, testConfig.Password, testConfig.DBName)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		t.Fatal("failed to connect to the database:", err)
	}
	defer db.Close()

	router := mux.NewRouter()

	// Endpointy
	router.HandleFunc("/api/News", GetAllNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("GET")
	router.HandleFunc("/api/News/{id}", GetNewsByID(db, testConfig.SchemaName, testConfig.TableName)).Methods("GET")
	router.HandleFunc("/api/News", CreateNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("POST")
	router.HandleFunc("/api/News/{id}", UpdateNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("PUT")
	router.HandleFunc("/api/News/{id}", DeleteNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("DELETE")

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		t.Fatal("cannot run server on port 8080:", err)
	}
}

//Test GetAllNews method for api/News endpoint
func TestGetAllNews(t *testing.T) {
	db, testConfig, err := RunTestingServer()
	if err != nil {
		t.Fatal("cannot run server on port 8080:", err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal("cannot connect:", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal("cannot close DB connection:", err)
		}
	}()

	//nowy request
	req := httptest.NewRequest(http.MethodGet, "/api/News", nil)
	recorder := httptest.NewRecorder()

	handler := GetAllNews(db, testConfig.SchemaName, testConfig.TableName)
	handler.ServeHTTP(recorder, req)

	//weryfikacja stanu kodu
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	err = ShutdownTestingServer()
	if err != nil {
		t.Fatal("error shutting down server:", err)
	}
}

//Test GetNewsByID method for api/News/{id} endpoint
func TestGetNewsByID(t *testing.T) {
	db, testConfig, err := RunTestingServer()
	if err != nil {
		t.Fatal("cannot run server on port 8080:", err)
	}
	if err = db.Ping(); err != nil {
		t.Fatal("cannot connect:", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/News/{id}", GetNewsByID(db, testConfig.SchemaName, testConfig.TableName)).Methods("GET")

	// Test 1: Poprawny ID
	req1 := httptest.NewRequest(http.MethodGet, "/api/News/1", nil)
	recorder1 := httptest.NewRecorder()
	router.ServeHTTP(recorder1, req1)
	if recorder1.Code != http.StatusOK {
		t.Errorf("ID=1, expected status code %d, got %d", http.StatusOK, recorder1.Code)
	}

	// Test 2: Niepoprawny ID
	req2 := httptest.NewRequest(http.MethodGet, "/api/News/invalid", nil)
	recorder2 := httptest.NewRecorder()
	router.ServeHTTP(recorder2, req2)
	if recorder2.Code != http.StatusBadRequest {
		t.Errorf("ID=invalid, expected status code %d, got %d", http.StatusBadRequest, recorder2.Code)
	}

	// Test 3: Nie istniejący ID
	req3 := httptest.NewRequest(http.MethodGet, "/api/News/999", nil)
	recorder3 := httptest.NewRecorder()
	router.ServeHTTP(recorder3, req3)
	if recorder3.Code != http.StatusNotFound {
		t.Errorf("ID=999, expected status code %d, got %d", http.StatusNotFound, recorder3.Code)
	}
}

//Test CreateNews method for api/News endpoint
func TestCreateNews(t *testing.T) {
	tests := []struct {
		RequestBody    interface{}
		Token          string
		ExpectedStatus int
	}{
		{
			RequestBody:    map[string]string{"content": "Test news content"},
			Token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwOi8vc2NoZW1hcy54bWxzb2FwLm9yZy93cy8yMDA1LzA1L2lkZW50aXR5L2NsYWltcy9uYW1lIjoidGVzdCIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL25hbWVpZGVudGlmaWVyIjoiMzU1OWIzNDktZWY1NS00MDQwLWE5ZjgtYjFhYzAwNWE1YzkxIiwibmJmIjoiMTY4ODA0NjI3MCIsImV4cCI6IjE2ODgxMzI2NzAiLCJodHRwOi8vc2NoZW1hcy5taWNyb3NvZnQuY29tL3dzLzIwMDgvMDYvaWRlbnRpdHkvY2xhaW1zL3JvbGUiOiJhZG1pbiJ9.NjZ6A8drgYDqlVWwjrSbA07vcuBQuPwhBnt0Zv2l-eE",
			ExpectedStatus: http.StatusCreated,
		},
		{
			RequestBody:    map[string]string{"content": ""},
			Token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwOi8vc2NoZW1hcy54bWxzb2FwLm9yZy93cy8yMDA1LzA1L2lkZW50aXR5L2NsYWltcy9uYW1lIjoidGVzdCIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL25hbWVpZGVudGlmaWVyIjoiMzU1OWIzNDktZWY1NS00MDQwLWE5ZjgtYjFhYzAwNWE1YzkxIiwibmJmIjoiMTY4ODA0NjI3MCIsImV4cCI6IjE2ODgxMzI2NzAiLCJodHRwOi8vc2NoZW1hcy5taWNyb3NvZnQuY29tL3dzLzIwMDgvMDYvaWRlbnRpdHkvY2xhaW1zL3JvbGUiOiJhZG1pbiJ9.NjZ6A8drgYDqlVWwjrSbA07vcuBQuPwhBnt0Zv2l-eE",
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	db, testConfig, err := RunTestingServer()
	if err != nil {
		t.Fatal("cannot run server on port 8888:", err)
	}
	defer db.Close()

	for _, tc := range tests {
		jsonData, err := json.Marshal(tc.RequestBody)
		if err != nil {
			t.Fatal("failed to marshal request body to JSON:", err)
		}

		//umieszczenie tokena JWT wewnątrz elementu Body zapytania
		req := httptest.NewRequest(http.MethodPost, "/api/News", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+tc.Token)

		recorder := httptest.NewRecorder()
		handler := CreateNews(db, testConfig.SchemaName, testConfig.TableName)
		handler.ServeHTTP(recorder, req)

		if recorder.Code != tc.ExpectedStatus {
			t.Errorf("Expected status code %d, got %d", tc.ExpectedStatus, recorder.Code)
		}
	}
}

//Test UpdateNews method for api/News/{id} endpoint
func TestUpdateNews(t *testing.T) {
	tests := []struct {
		NewsID         string
		RequestBody    interface{}
		Token          string
		ExpectedStatus int
	}{
		{
			NewsID:         "1",
			RequestBody:    map[string]string{"content": "Updated news content"},
			Token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwOi8vc2NoZW1hcy54bWxzb2FwLm9yZy93cy8yMDA1LzA1L2lkZW50aXR5L2NsYWltcy9uYW1lIjoidGVzdCIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL25hbWVpZGVudGlmaWVyIjoiMzU1OWIzNDktZWY1NS00MDQwLWE5ZjgtYjFhYzAwNWE1YzkxIiwibmJmIjoiMTY4ODA0NjI3MCIsImV4cCI6IjE2ODgxMzI2NzAiLCJodHRwOi8vc2NoZW1hcy5taWNyb3NvZnQuY29tL3dzLzIwMDgvMDYvaWRlbnRpdHkvY2xhaW1zL3JvbGUiOiJhZG1pbiJ9.NjZ6A8drgYDqlVWwjrSbA07vcuBQuPwhBnt0Zv2l-eE",
			ExpectedStatus: http.StatusOK,
		},
		{
			NewsID:         "999",
			RequestBody:    map[string]string{"content": "Updated news content"},
			Token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwOi8vc2NoZW1hcy54bWxzb2FwLm9yZy93cy8yMDA1LzA1L2lkZW50aXR5L2NsYWltcy9uYW1lIjoidGVzdCIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL25hbWVpZGVudGlmaWVyIjoiMzU1OWIzNDktZWY1NS00MDQwLWE5ZjgtYjFhYzAwNWE1YzkxIiwibmJmIjoiMTY4ODA0NjI3MCIsImV4cCI6IjE2ODgxMzI2NzAiLCJodHRwOi8vc2NoZW1hcy5taWNyb3NvZnQuY29tL3dzLzIwMDgvMDYvaWRlbnRpdHkvY2xhaW1zL3JvbGUiOiJhZG1pbiJ9.NjZ6A8drgYDqlVWwjrSbA07vcuBQuPwhBnt0Zv2l-eE",
			ExpectedStatus: http.StatusOK,
		},
	}

	db, testConfig, err := RunTestingServer()
	if err != nil {
		t.Fatal("cannot run server on port 8888:", err)
	}
	defer db.Close()

	//nowy router do obsługi ządań serwera
	router := mux.NewRouter()
	router.HandleFunc("/api/News/{id}", UpdateNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("PUT")

	for _, tc := range tests {
		jsonData, err := json.Marshal(tc.RequestBody)
		if err != nil {
			t.Fatal("failed to marshal request body to JSON:", err)
		}

		//tworzenie zapytania
		req := httptest.NewRequest(http.MethodPut, "/api/News/"+tc.NewsID, bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+tc.Token)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != tc.ExpectedStatus {
			t.Errorf("Expected status code %d, got %d", tc.ExpectedStatus, recorder.Code)
		}
	}
}

//Test DeleteNews method for api/News/{id} endpoint
func TestDeleteNews(t *testing.T) {
	tests := []struct {
		NewsID         string
		Token          string
		ExpectedStatus int
	}{
		{
			NewsID:         "1",
			Token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwOi8vc2NoZW1hcy54bWxzb2FwLm9yZy93cy8yMDA1LzA1L2lkZW50aXR5L2NsYWltcy9uYW1lIjoidGVzdCIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL25hbWVpZGVudGlmaWVyIjoiMzU1OWIzNDktZWY1NS00MDQwLWE5ZjgtYjFhYzAwNWE1YzkxIiwibmJmIjoiMTY4ODA0NjI3MCIsImV4cCI6IjE2ODgxMzI2NzAiLCJodHRwOi8vc2NoZW1hcy5taWNyb3NvZnQuY29tL3dzLzIwMDgvMDYvaWRlbnRpdHkvY2xhaW1zL3JvbGUiOiJhZG1pbiJ9.NjZ6A8drgYDqlVWwjrSbA07vcuBQuPwhBnt0Zv2l-eE",
			ExpectedStatus: http.StatusOK,
		},
		{
			NewsID:         "666",
			Token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwOi8vc2NoZW1hcy54bWxzb2FwLm9yZy93cy8yMDA1LzA1L2lkZW50aXR5L2NsYWltcy9uYW1lIjoidGVzdCIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL25hbWVpZGVudGlmaWVyIjoiMzU1OWIzNDktZWY1NS00MDQwLWE5ZjgtYjFhYzAwNWE1YzkxIiwibmJmIjoiMTY4ODA0NjI3MCIsImV4cCI6IjE2ODgxMzI2NzAiLCJodHRwOi8vc2NoZW1hcy5taWNyb3NvZnQuY29tL3dzLzIwMDgvMDYvaWRlbnRpdHkvY2xhaW1zL3JvbGUiOiJhZG1pbiJ9.NjZ6A8drgYDqlVWwjrSbA07vcuBQuPwhBnt0Zv2l-eE",
			ExpectedStatus: http.StatusNotFound,
		},
	}

	db, testConfig, err := RunTestingServer()
	if err != nil {
		t.Fatal("cannot run server on port 8080:", err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/api/News/{id}", DeleteNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("DELETE")

	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodDelete, "/api/News/"+tc.NewsID, nil)
		req.Header.Set("Authorization", "Bearer "+tc.Token)

		//śledzi stan zapytan http
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != tc.ExpectedStatus {
			t.Errorf("Expected status code %d, got %d", tc.ExpectedStatus, recorder.Code)
		}
	}
}

//tworzenie serwera dla funkcji testujących
func RunTestingServer() (*sql.DB, config.Config, error) {
	var testConfig config.Config
	var db *sql.DB
	configData, err := ioutil.ReadFile("../config/testConfig.json")
	if err != nil {
		return db, testConfig, err
	}

	//odczytuje config
	err = json.Unmarshal(configData, &testConfig)
	if err != nil {
		return db, testConfig, err
	}

	//łączenie z bazą
	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		testConfig.Host, testConfig.Port, testConfig.User, testConfig.Password, testConfig.DBName)
	db, err = sql.Open("postgres", dbInfo)
	if err != nil {
		return db, testConfig, err
	}

	router := mux.NewRouter()

	// Endpointy
	router.HandleFunc("/api/News", GetAllNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("GET")
	router.HandleFunc("/api/News/{id}", GetNewsByID(db, testConfig.SchemaName, testConfig.TableName)).Methods("GET")
	router.HandleFunc("/api/News", CreateNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("POST")
	router.HandleFunc("/api/News/{id}", UpdateNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("PUT")
	router.HandleFunc("/api/News/{id}", DeleteNews(db, testConfig.SchemaName, testConfig.TableName)).Methods("DELETE")

	serverInstance = &http.Server{Addr: ":8080", Handler: router}
	go func() {
		err := serverInstance.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Println("error while starting server:", err)
		}
	}()

	return db, testConfig, nil
}

//zamknięcie instancji serwera
func ShutdownTestingServer() error {
	if serverInstance != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return serverInstance.Shutdown(ctx)
	}
	return nil
}
