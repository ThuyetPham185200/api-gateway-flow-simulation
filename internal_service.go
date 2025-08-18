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

	// 1️⃣ Auth Service
	authService := InternalService{
		Name: "AuthService",
		Port: 9090,
		Endpoints: map[string]http.HandlerFunc{
			"/login": func(w http.ResponseWriter, r *http.Request) {
				logRequest("AuthService", r)
				resp := map[string]string{
					"token": "fake-jwt-token-" + time.Now().Format("150405"),
				}
				writeJSON(w, resp)
			},
			"/register": func(w http.ResponseWriter, r *http.Request) {
				logRequest("AuthService", r)
				var body map[string]string
				_ = json.NewDecoder(r.Body).Decode(&body)
				resp := map[string]string{
					"user_id":  "user-" + time.Now().Format("150405"),
					"username": body["username"],
					"status":   "registered",
				}
				writeJSON(w, resp)
			},
			"/me/password": func(w http.ResponseWriter, r *http.Request) {
				logRequest("AuthService", r)
				resp := map[string]string{
					"message": "Password updated",
				}
				writeJSON(w, resp)
			},
			"/me": func(w http.ResponseWriter, r *http.Request) {
				logRequest("AuthService", r)
				resp := map[string]string{
					"message": "Account soft deleted",
				}
				writeJSON(w, resp)
			},
		},
	}

	// 2️⃣ User/Profile Service
	userService := InternalService{
		Name: "UserService",
		Port: 9091,
		Endpoints: map[string]http.HandlerFunc{
			"/users/{user_id}": func(w http.ResponseWriter, r *http.Request) {
				logRequest("UserService", r)
				resp := map[string]interface{}{
					"user_id":   "user-123",
					"username":  "alice",
					"avatar":    "avatar.png",
					"bio":       "Hello world",
					"createdAt": time.Now().Format(time.RFC3339),
				}
				writeJSON(w, resp)
			},
			"/me": func(w http.ResponseWriter, r *http.Request) {
				logRequest("UserService", r)
				var body map[string]string
				_ = json.NewDecoder(r.Body).Decode(&body)
				resp := map[string]string{
					"message":  "Profile updated",
					"username": body["username"],
					"bio":      body["bio"],
				}
				writeJSON(w, resp)
			},
			"/users": func(w http.ResponseWriter, r *http.Request) {
				logRequest("UserService", r)
				resp := map[string]interface{}{
					"users": []map[string]string{
						{"user_id": "user-123", "username": "alice", "avatar": "avatar.png"},
						{"user_id": "user-456", "username": "bob", "avatar": "bob.png"},
					},
					"total": 2,
				}
				writeJSON(w, resp)
			},
		},
	}

	// 3️⃣ Posts Service
	postsService := InternalService{
		Name: "PostsService",
		Port: 9092,
		Endpoints: map[string]http.HandlerFunc{
			"/posts/{post_id}": func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					// Handle Get Post
					logRequest("PostsService [GET]", r)
					resp := map[string]interface{}{
						"post_id":   "post-123",
						"user_id":   "user-123",
						"content":   "Hello World",
						"createdAt": time.Now(),
						"media_ids": []int{1, 2},
					}
					writeJSON(w, resp)
				case http.MethodPatch:
					// Handle Update Post
					logRequest("PostsService [PATCH]", r)
					var body map[string]interface{}
					_ = json.NewDecoder(r.Body).Decode(&body)
					resp := map[string]interface{}{
						"message": "Post updated",
						"post_id": "post-123",
						"updated": body,
					}
					writeJSON(w, resp)
				case http.MethodDelete:
					// Handle Delete Post
					logRequest("PostsService [DELETE]", r)
					resp := map[string]interface{}{
						"message": "Post soft deleted",
						"post_id": "post-123",
					}
					writeJSON(w, resp)
				default:
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			},

			"/users/{user_id}/posts": func(w http.ResponseWriter, r *http.Request) {
				// Handle Get User Posts
				logRequest("PostsService [User Posts]", r)
				resp := map[string]interface{}{
					"posts": []map[string]interface{}{
						{"post_id": "post-101", "content": "User post 1", "createdAt": time.Now()},
						{"post_id": "post-102", "content": "User post 2", "createdAt": time.Now()},
					},
					"total": 2,
				}
				writeJSON(w, resp)
			},

			"/me/posts": func(w http.ResponseWriter, r *http.Request) {
				// Handle Get Own Posts
				logRequest("PostsService [Own Posts]", r)
				resp := map[string]interface{}{
					"posts": []map[string]interface{}{
						{"post_id": "post-201", "content": "My post 1", "createdAt": time.Now()},
						{"post_id": "post-202", "content": "My post 2", "createdAt": time.Now()},
					},
					"total": 2,
				}
				writeJSON(w, resp)
			},

			"/posts": func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					// Handle Create Post
					logRequest("PostsService [Create]", r)
					var body map[string]interface{}
					_ = json.NewDecoder(r.Body).Decode(&body)
					resp := map[string]interface{}{
						"post_id": "post-new",
						"message": "Post created",
						"data":    body,
					}
					writeJSON(w, resp)
				} else {
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			},
		},
	}

	// 4️⃣ Reactions Service
	reactionsService := InternalService{
		Name: "ReactionsService",
		Port: 9093,
		Endpoints: map[string]http.HandlerFunc{
			"/posts/{post_id}/reactions": func(w http.ResponseWriter, r *http.Request) {
				logRequest("ReactionsService", r)
				switch r.Method {
				case http.MethodGet:
					resp := map[string]interface{}{
						"count": 3,
						"types": []string{"like", "love"},
						"users": []map[string]string{
							{"user_id": "user-123", "username": "alice"},
						},
						"total": 1,
					}
					writeJSON(w, resp)
				case http.MethodPost:
					var body map[string]string
					_ = json.NewDecoder(r.Body).Decode(&body)
					writeJSON(w, map[string]string{"message": "Reaction added"})
				case http.MethodDelete:
					var body map[string]string
					_ = json.NewDecoder(r.Body).Decode(&body)
					writeJSON(w, map[string]string{"message": "Reaction removed"})
				}
			},
		},
	}

	// ===== Khởi chạy các service song song =====
	services := []*InternalService{&authService, &userService, &postsService, &reactionsService}
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
