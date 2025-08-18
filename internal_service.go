// internal_service.go
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	// ===== /auth/login =====
	http.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		resp := map[string]string{
			"token": "fake-jwt-token-" + time.Now().Format("150405"),
		}
		writeJSON(w, resp)
	})

	// ===== /auth/register =====
	http.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		resp := map[string]string{
			"user_id":  "user-" + time.Now().Format("150405"),
			"username": body["login"],
			"status":   "registered",
		}
		writeJSON(w, resp)
	})

	// ===== /profile/get =====
	http.HandleFunc("/profile/get", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		resp := map[string]string{
			"user_id":  "user-123",
			"username": "alice",
			"email":    "alice@example.com",
		}
		writeJSON(w, resp)
	})

	// ===== /profile/update =====
	http.HandleFunc("/profile/update", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		resp := map[string]string{
			"status":   "updated",
			"username": body["username"],
			"email":    body["email"],
		}
		writeJSON(w, resp)
	})

	log.Println("🟢 Internal service running at :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

// ===== Helpers =====

// Log request nhận được
func logRequest(r *http.Request) {
	log.Println("====== Internal Service Received ======")
	log.Printf("Method: %s URL: %s", r.Method, r.URL.String())
	for k, v := range r.Header {
		log.Printf("%s: %v", k, v)
	}
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	log.Println("Body:", string(body))
	log.Println("======================================")
	// Restore body để đọc lại nếu cần downstream
	r.Body = io.NopCloser(bytes.NewReader(body))
}

// Ghi JSON response
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
