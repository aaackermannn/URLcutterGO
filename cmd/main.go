package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// Модели данных
type URL struct {
	ID        string    `json:"id"`
	Original  string    `json:"original_url"`
	Short     string    `json:"short_url"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int       `json:"clicks"`
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	ShortURL string `json:"short_url"`
}

func main() {
	// Подключение к MySQL
	connStr := "urluser:password@tcp(localhost:3306)/url_shortener?parseTime=true"
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("✅ Successfully connected to MySQL database")

	// Создание таблицы
	if err := createTable(db); err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// Настройка роутера
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/shorten", createShortURLHandler(db)).Methods("POST")
	api.HandleFunc("/url/{short}", getURLInfoHandler(db)).Methods("GET")

	// Redirect route
	r.HandleFunc("/{short}", redirectHandler(db)).Methods("GET")

	// Health check
	r.HandleFunc("/health", healthHandler).Methods("GET")

	// Serve frontend files
	r.PathPrefix("/").Handler(serveFrontend())

	log.Println("🚀 Server starting on :8080")
	log.Println("🌐 Frontend available at http://localhost:8080")
	log.Println("📋 API endpoints:")
	log.Println("   POST /api/v1/shorten")
	log.Println("   GET  /api/v1/url/{short}")
	log.Println("   GET  /{short}")
	log.Println("   GET  /health")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func createTable(db *sql.DB) error {
	// Создаем таблицу
	createTableQuery := `CREATE TABLE IF NOT EXISTS urls (
		id VARCHAR(10) PRIMARY KEY,
		original_url TEXT NOT NULL,
		short_url VARCHAR(10) UNIQUE NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		clicks INT DEFAULT 0
	)`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Создаем индексы отдельными командами
	indexQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_short_url ON urls(short_url)",
		"CREATE INDEX IF NOT EXISTS idx_original_url ON urls(original_url(255))",
	}

	for i, query := range indexQueries {
		_, err = db.Exec(query)
		if err != nil {
			log.Printf("Warning: Failed to create index %d: %v", i+1, err)
		}
	}

	log.Println("✅ Database table created successfully")
	return nil
}

func serveFrontend() http.Handler {
	if _, err := os.Stat("web"); os.IsNotExist(err) {
		log.Println("⚠️  Frontend folder 'web' not found, serving API only")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(`
					<!DOCTYPE html>
					<html>
					<head>
						<title>URL Shortener API</title>
						<style>
							body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
							.endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
							.method { font-weight: bold; color: #007bff; }
						</style>
					</head>
					<body>
						<h1>🔗 URL Shortener API</h1>
						<p>API сервер работает. Для веб-интерфейса создайте папку 'web' с файлами фронтенда.</p>
						<div class="endpoint">
							<span class="method">POST</span> /api/v1/shorten - Создать короткую ссылку
						</div>
						<div class="endpoint">
							<span class="method">GET</span> /api/v1/url/{short} - Получить информацию о ссылке
						</div>
						<div class="endpoint">
							<span class="method">GET</span> /{short} - Перенаправление на оригинальный URL
						</div>
						<div class="endpoint">
							<span class="method">GET</span> /health - Проверка здоровья сервиса
						</div>
					</body>
					</html>
				`))
			} else {
				http.NotFound(w, r)
			}
		})
	}

	log.Println("🌐 Serving frontend from 'web' folder")
	return http.FileServer(http.Dir("web"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func createShortURLHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateURLRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		// Проверяем, не сокращали ли уже этот URL
		var existingURL URL
		err := db.QueryRow("SELECT short_url FROM urls WHERE original_url = ?", req.URL).Scan(&existingURL.Short)
		if err == nil {
			// URL уже существует, возвращаем существующую короткую ссылку
			response := CreateURLResponse{ShortURL: existingURL.Short}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Генерируем короткую ссылку
		shortURL := generateShortURL()

		// Сохраняем в базу данных
		_, err = db.Exec(
			"INSERT INTO urls (id, original_url, short_url, created_at, clicks) VALUES (?, ?, ?, ?, ?)",
			shortURL, req.URL, shortURL, time.Now(), 0,
		)

		if err != nil {
			log.Printf("Error creating short URL: %v", err)
			http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
			return
		}

		response := CreateURLResponse{ShortURL: shortURL}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func getURLInfoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		short := vars["short"]

		var url URL
		err := db.QueryRow(
			"SELECT id, original_url, short_url, created_at, clicks FROM urls WHERE short_url = ?",
			short,
		).Scan(&url.ID, &url.Original, &url.Short, &url.CreatedAt, &url.Clicks)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "URL not found", http.StatusNotFound)
				return
			}
			log.Printf("Error fetching URL info: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(url)
	}
}

func redirectHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		short := vars["short"]

		var originalURL string
		err := db.QueryRow("SELECT original_url FROM urls WHERE short_url = ?", short).Scan(&originalURL)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "URL not found", http.StatusNotFound)
				return
			}
			log.Printf("Error redirecting: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Увеличиваем счетчик кликов
		go func() {
			_, err := db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE short_url = ?", short)
			if err != nil {
				log.Printf("Error updating click count: %v", err)
			}
		}()

		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}

func generateShortURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	short := make([]byte, length)
	for i := range short {
		short[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(short)
}
