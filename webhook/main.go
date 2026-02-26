package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// 用环境变量传入 secret，避免硬编码
var secret = os.Getenv("WEBHOOK_SECRET")

func verify(signature string, body []byte) bool {
	if secret == "" {
		return true // 没设 secret 就跳过验证
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func handler(w http.ResponseWriter, r *http.Request) {
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

	cmd := exec.Command("bash", "-c", "cd /home/ubuntu/blog && git pull && /snap/bin/hugo")
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

func main() {
	http.HandleFunc("/webhook", handler)
	log.Println("webhook server listening on :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
