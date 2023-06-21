package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

const (
	schemaName = "newsService"
	tableName  = "News"
)

type News struct {
	ID          int    `json:"id" db:"Id"`
	Content     string `json:"content" db:"content"`
	CreatedDate string `json:"createdDate" db:"createdDate"`
	AuthorID    int    `json:"authorId" db:"authorId"`
	LastUpdate  string `json:"lastUpdate" db:"lastUpdate"`
}

type NewNews struct {
	Content string `json:"content"`
}

type LoginCredentials struct {
	ID        int    `json:"ID"`
	GrantType string `json:"grant_type"`
}

func GetAllNews(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Wykonanie zapytania SELECT
		query := fmt.Sprintf(`SELECT * FROM "%s"."%s"`, schemaName, tableName)
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Przetwarzanie wyników
		newsList := make([]News, 0)
		for rows.Next() {
			var news News
			err := rows.Scan(&news.ID, &news.Content, &news.CreatedDate, &news.AuthorID, &news.LastUpdate)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			newsList = append(newsList, news)
		}

		// Konwersja do formatu JSON
		jsonData, err := json.Marshal(newsList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Ustawienie nagłówka i zwrócenie odpowiedzi
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func GetNewsByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pobranie wartości parametru "id" z ścieżki
		vars := mux.Vars(r)
		idStr := vars["id"]

		// Konwersja parametru "id" na int
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid news ID", http.StatusBadRequest)
			return
		}

		// Wykonanie zapytania SELECT z mapowaniem nazw kolumn
		query := fmt.Sprintf(`SELECT "Id", "Content", "CreatedDate", "AuthorId", "LastUpdate" FROM "%s"."%s" WHERE "Id"=$1`, schemaName, tableName)
		row := db.QueryRow(query, id)

		// Przetworzenie wyniku
		var news News
		err = row.Scan(&news.ID, &news.Content, &news.CreatedDate, &news.AuthorID, &news.LastUpdate)
		if err == sql.ErrNoRows {
			http.Error(w, "News not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Konwersja do formatu JSON
		jsonData, err := json.Marshal(news)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Ustawienie nagłówka i zwrócenie odpowiedzi
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func CreateNews(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Sprawdzenie uprawnień użytkownika na podstawie tokenu JWT w nagłówku Authorization
		tokenString := r.Header.Get("Authorization")
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		fmt.Println("Token: ", tokenString)
		claims, err := validateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			fmt.Println("Token error: ", err)
			return
		}

		// odczytujemy role z tokena
		if claims.GrantType != "admin" && claims.GrantType != "employee" {
			fmt.Println("Brak roli admin lub employee")
			return
		}
		authorID := claims.ID

		// Odczytanie danych nowego news'a z ciała żądania
		var newNews NewNews
		err = json.NewDecoder(r.Body).Decode(&newNews)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Sprawdzenie, czy treść news'a jest niepusta
		if newNews.Content == "" {
			http.Error(w, "News content cannot be empty", http.StatusBadRequest)
			return
		}

		// Wstawienie nowego news'a do bazy danych
		query := fmt.Sprintf(`INSERT INTO "%s"."%s" ("Content", "CreatedDate", "AuthorId", "LastUpdate") VALUES ($1, NOW(), $2, NOW()) RETURNING "Id"`, schemaName, tableName)
		var newsID int
		err = db.QueryRow(query, newNews.Content, authorID).Scan(&newsID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Utworzenie odpowiedzi zawierającej ID utworzonego news'a
		response := map[string]int{"id": newsID}
		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Ustawienie nagłówka i zwrócenie odpowiedzi
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonData)
	}
}

func UpdateNews(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pobranie tokena z nagłówka Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Brak tokena uwierzytelniającego", http.StatusUnauthorized)
			return
		}

		// Wyodrębnienie samego tokenu bez prefiksu "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parsowanie i weryfikacja tokenu
		claims, err := validateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			fmt.Println("Token error: ", err)
			return
		}

		if claims.GrantType != "admin" && claims.GrantType != "employee" {
			fmt.Println("Brak roli admin lub employee")
			return
		}
		// Pobranie identyfikatora newsa z parametru ścieżki
		vars := mux.Vars(r)
		newsID := vars["id"]

		// Odczytanie treści newsa z ciała żądania
		var newsData struct {
			Content string `json:"content"`
		}
		err = json.NewDecoder(r.Body).Decode(&newsData)
		if err != nil {
			http.Error(w, "Błąd odczytu danych żądania", http.StatusBadRequest)
			return
		}

		// Aktualizacja newsa w bazie danych
		query := `UPDATE "newsService"."News" SET "Content"=$1, "LastUpdate"=NOW() WHERE "Id"=$2`
		_, err = db.Exec(query, newsData.Content, newsID)
		if err != nil {
			http.Error(w, "Błąd podczas aktualizacji newsa", http.StatusInternalServerError)
			return
		}

		// Zwrócenie odpowiedzi sukcesu
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("News został zaktualizowany"))
	}
}

func DeleteNews(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pobranie tokena z nagłówka Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Brak tokena uwierzytelniającego", http.StatusUnauthorized)
			return
		}

		// Wyodrębnienie samego tokenu bez prefiksu "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := validateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			fmt.Println("Token error: ", err)
			return
		}

		if claims.GrantType != "admin" && claims.GrantType != "employee" {
			fmt.Println("Brak roli admin lub employee")
			return
		}

		// Pobranie identyfikatora newsa z parametru ścieżki
		vars := mux.Vars(r)
		newsID := vars["id"]

		// Usunięcie newsa z bazy danych
		// ...
		query := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE "Id"=$1`, schemaName, tableName)
		result, err := db.Exec(query, newsID)
		if err != nil {
			http.Error(w, "Błąd podczas usuwania newsa z bazy danych", http.StatusInternalServerError)
			return
		}

		// Sprawdzenie liczby usuniętych wierszy
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, "Błąd podczas pobierania liczby usuniętych wierszy", http.StatusInternalServerError)
			return
		}

		// Sprawdzenie, czy jakikolwiek wiersz został usunięty
		if rowsAffected == 0 {
			http.Error(w, "Nie znaleziono newsa o podanym identyfikatorze", http.StatusNotFound)
			return
		}

		// Zwrócenie odpowiedzi sukcesu
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("News został usunięty"))
	}
}

// Funkcja do weryfikacji tokenu JWT i odczytu informacji
func validateToken(tokenString string) (*LoginCredentials, error) {
	// Parsowanie tokena JWT
	// Parsowanie i weryfikacja tokenu
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Sprawdzenie algorytmu podpisu tokenu (HMAC)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("nieprawidłowy algorytm podpisu: %v", token.Header["alg"])
		}

		// Klucz tajny używany do weryfikacji tokenu (ten sam klucz, który został użyty do podpisania tokenu)
		secretKey := []byte("MySecretKeyIsSecretSoDoNotTell")

		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("nieznany blad parsowania")
	}

	var loginCredentials LoginCredentials
	// Sprawdzenie ważności tokenu
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		numStr := claims["id"].(string)
		num, err := strconv.Atoi(numStr)
		if err != nil {
			fmt.Println("Błąd konwersji:", err)
			return nil, err
		}
		loginCredentials.ID = num
		fmt.Println("ID:", loginCredentials.ID)
		loginCredentials.GrantType = claims["grant_type"].(string)
		fmt.Println("Grant Type:", loginCredentials.GrantType)

	} else {
		fmt.Println("Nieprawidłowy format tokenu")
		return nil, err
	}
	return &loginCredentials, err
}
