package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	_ "github.com/mattn/go-sqlite3"
)

var secret = os.Getenv("WEBHOOK_SECRET")

var db *sql.DB

func initDB() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/home/ubuntu/blog-views.db"
	}
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS views (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("create table: %v", err)
	}
	// 按 path 建索引，加速查询
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_views_path ON views(path)`)
}

func verify(signature string, body []byte) bool {
	if secret == "" {
		return true
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	if !verify(r.Header.Get("X-Hub-Signature-256"), body) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	log.Println("received push event, rebuilding...")

	cmd := exec.Command("bash", "-c", "cd /home/ubuntu/blog && git pull && /usr/local/bin/hugo")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("build failed: %v", err)
		http.Error(w, "build failed", http.StatusInternalServerError)
		return
	}

	log.Println("build success")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	// 允许博客页面跨域请求
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		_, err := db.Exec("INSERT INTO views (path) VALUES (?)", path)
		if err != nil {
			log.Printf("insert view: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

	case http.MethodGet:
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM views WHERE path = ?", path).Scan(&count)
		if err != nil {
			log.Printf("query view: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"views": count})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/api/view", viewHandler)

	log.Println("webhook server listening on :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
