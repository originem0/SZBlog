package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var secret = os.Getenv("WEBHOOK_SECRET")

var db *sql.DB

// buildMu 串行化构建：避免两个 push（GitHub 重试 / 手动触发）并发跑
// git reset --hard + hugo，在同一目录互相踩导致 public/ 写出半成品。
var buildMu sync.Mutex

// webhook 请求体上限：HMAC 校验需读完整个 body，限制大小防内存耗尽。
const maxWebhookBody = 10 << 20 // 10 MiB

// rate limiter: 每 IP 每分钟最多 30 次 POST
var (
	rateMu     sync.Mutex
	rateMap    = make(map[string][]time.Time)
	rateLimit  = 30
	rateWindow = time.Minute
)

// path 校验：以 / 开头，长度 <= 200，只允许 URL 安全字符
var validPath = regexp.MustCompile(`^/[a-zA-Z0-9\-._~/]+$`)

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

	// 聚合计数表替代逐行记录
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS view_counts (
		path TEXT PRIMARY KEY,
		count INTEGER NOT NULL DEFAULT 0
	)`)
	if err != nil {
		log.Fatalf("create table: %v", err)
	}

	// 迁移：如果旧 views 表存在，把数据聚合到 view_counts
	migrateOldViews()
}

func migrateOldViews() {
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='views'").Scan(&tableName)
	if err != nil {
		return // 旧表不存在，无需迁移
	}

	log.Println("migrating old views table to view_counts...")
	_, err = db.Exec(`INSERT OR IGNORE INTO view_counts (path, count)
		SELECT path, COUNT(*) FROM views GROUP BY path`)
	if err != nil {
		log.Printf("migration error: %v", err)
		return
	}
	// 迁移成功后删除旧表
	db.Exec("DROP TABLE views")
	log.Println("migration complete, old views table dropped")
}

func verify(signature string, body []byte) bool {
	if secret == "" {
		log.Println("WEBHOOK_SECRET not set, rejecting request")
		return false
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

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBody))
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	if !verify(r.Header.Get("X-Hub-Signature-256"), body) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	// 串行化构建，防止并发 push 互相踩
	buildMu.Lock()
	defer buildMu.Unlock()

	log.Println("received push event, rebuilding...")

	// fetch + reset 比 pull 更可靠：避免 merge 冲突和脏工作区问题
	// --gc --cleanDestinationDir：清理 public/ 中已删除/重命名内容遗留的孤儿文件
	cmd := exec.Command("bash", "-c",
		"cd /home/ubuntu/blog && git fetch origin && git reset --hard origin/main && /usr/local/bin/hugo --gc --cleanDestinationDir")
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

// checkRateLimit 检查 IP 是否超出频率限制
func checkRateLimit(ip string) bool {
	rateMu.Lock()
	defer rateMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rateWindow)

	// 清理过期记录
	times := rateMap[ip]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rateLimit {
		rateMap[ip] = valid
		return false
	}

	rateMap[ip] = append(valid, now)
	return true
}

// clientIP 提取真实客户端 IP。nginx 用 `proxy_set_header X-Real-IP $remote_addr`
// 覆盖客户端传入值，故 X-Real-IP 可信。这里故意不读 X-Forwarded-For：本项目
// 的 nginx /api/ 块不设置该头，客户端传入的 XFF 会原样透传、可被伪造，用它做
// 限流 key 等于没限流。直连 :9000（防火墙限 GitHub IP）时退回连接地址去端口。
func clientIP(r *http.Request) string {
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// sweepRateMap 周期清理 rateMap：删除窗口内已无有效记录的 IP，
// 防止 map 单调增长（checkRateLimit 只增不删会导致内存泄漏）。
func sweepRateMap() {
	ticker := time.NewTicker(rateWindow)
	for range ticker.C {
		rateMu.Lock()
		cutoff := time.Now().Add(-rateWindow)
		for ip, times := range rateMap {
			valid := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rateMap, ip)
			} else {
				rateMap[ip] = valid
			}
		}
		rateMu.Unlock()
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	// CORS 限定到博客域名
	w.Header().Set("Access-Control-Allow-Origin", "https://sharonzhou.site")
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

	// path 校验
	if len(path) > 200 || !validPath.MatchString(path) || strings.Contains(path, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPost:
		// rate limit 检查
		if !checkRateLimit(clientIP(r)) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		_, err := db.Exec(
			`INSERT INTO view_counts (path, count) VALUES (?, 1)
			 ON CONFLICT(path) DO UPDATE SET count = count + 1`, path)
		if err != nil {
			log.Printf("upsert view: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

	case http.MethodGet:
		var count int
		err := db.QueryRow("SELECT count FROM view_counts WHERE path = ?", path).Scan(&count)
		if err == sql.ErrNoRows {
			count = 0
		} else if err != nil {
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

	go sweepRateMap()

	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/api/view", viewHandler)

	// 显式超时防慢连接耗尽（Slowloris）。WriteTimeout 设为 0：webhook 同步执行
	// git+hugo 构建，写响应前可能耗时较久，写超时会误伤成功构建的响应。
	// ReadHeaderTimeout 才是 Slowloris 的关键防护。
	srv := &http.Server{
		Addr:              ":9000",
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("webhook server listening on :9000")
	log.Fatal(srv.ListenAndServe())
}
