package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// ===== Struct Service =====
type InternalService struct {
	Name      string
	Port      int
	Endpoints map[string]http.HandlerFunc
}

func main() {
	// ===== Định nghĩa các service =====
	authService := InternalService{
		Name: "AuthService",
		Port: 9090,
		Endpoints: map[string]http.HandlerFunc{
			"/auth/login": func(w http.ResponseWriter, r *http.Request) {
				logRequest("AuthService", r)
				resp := map[string]string{
					"token": "fake-jwt-token-" + time.Now().Format("150405"),
				}
				writeJSON(w, resp)
			},
			"/auth/register": func(w http.ResponseWriter, r *http.Request) {
				logRequest("AuthService", r)
				var body map[string]string
				_ = json.NewDecoder(r.Body).Decode(&body)
				resp := map[string]string{
					"user_id":  "user-" + time.Now().Format("150405"),
					"username": body["login"],
					"status":   "registered",
				}
				writeJSON(w, resp)
			},
		},
	}

	profileService := InternalService{
		Name: "ProfileService",
		Port: 9091,
		Endpoints: map[string]http.HandlerFunc{
			"/profile/get": func(w http.ResponseWriter, r *http.Request) {
				logRequest("ProfileService", r)
				resp := map[string]string{
					"user_id":  "user-123",
					"username": "alice",
					"email":    "alice@example.com",
				}
				writeJSON(w, resp)
			},
		},
	}

	profileUpdateService := InternalService{
		Name: "ProfileUpdateService",
		Port: 9092,
		Endpoints: map[string]http.HandlerFunc{
			"/profile/update": func(w http.ResponseWriter, r *http.Request) {
				logRequest("ProfileUpdateService", r)
				var body map[string]string
				_ = json.NewDecoder(r.Body).Decode(&body)
				resp := map[string]string{
					"status":   "updated",
					"username": body["username"],
					"email":    body["email"],
				}
				writeJSON(w, resp)
			},
		},
	}

	// ===== Khởi chạy các service song song =====
	services := []*InternalService{&authService, &profileService, &profileUpdateService}
	for _, s := range services {
		go s.Start()
	}

	select {} // block chính
}

// ===== Method Start cho InternalService =====
func (s *InternalService) Start() {
	mux := http.NewServeMux()
	for path, handler := range s.Endpoints {
		mux.HandleFunc(path, handler)
	}
	addr := fmt.Sprintf(":%d", s.Port)
	log.Printf("🟢 %s running at %s", s.Name, addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// ===== Helpers =====
func logRequest(serviceName string, r *http.Request) {
	log.Printf("====== %s Received ======", serviceName)
	log.Printf("Method: %s URL: %s", r.Method, r.URL.String())
	for k, v := range r.Header {
		log.Printf("%s: %v", k, v)
	}
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	log.Println("Body:", string(body))
	log.Println("=================================")
	r.Body = io.NopCloser(bytes.NewReader(body))
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
