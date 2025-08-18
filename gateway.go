// gateway.go (updated)
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ===== Models =====

type RawRequestData struct {
	Ctx     context.Context
	Method  string
	Path    string
	Header  http.Header
	Body    []byte
	IP      string
	Topic   string
	Token   string
	ReplyCh chan GatewayResult
}

type GatewayResult struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// ===== Queue & Worker =====

var requestQueue = make(chan RawRequestData, 1024)

func startWorkers(n int) {
	for i := 0; i < n; i++ {
		go func(id int) {
			for req := range requestQueue {
				process(req)
			}
		}(i)
	}
}

// ===== Handler =====

func main() {
	startWorkers(4)

	// Các endpoint
	http.HandleFunc("/auth/login", makeHandler("auth/login"))
	http.HandleFunc("/auth/register", makeHandler("auth/register"))
	http.HandleFunc("/profile/get", makeHandler("profile/get"))
	http.HandleFunc("/profile/update", makeHandler("profile/update"))

	log.Println("🚀 API Gateway running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Helper tạo handler để tránh lặp code
func makeHandler(topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			body = b
		}

		replyCh := make(chan GatewayResult, 1)
		fmt.Println("================> %p", replyCh)
		job := RawRequestData{
			Ctx:     r.Context(),
			Method:  r.Method,
			Path:    r.URL.Path,
			Header:  r.Header.Clone(),
			Body:    body,
			IP:      r.RemoteAddr,
			Topic:   topic,
			Token:   r.Header.Get("Authorization"),
			ReplyCh: replyCh,
		}

		select {
		case requestQueue <- job:
		case <-r.Context().Done():
			writePlainError(w, http.StatusRequestTimeout, "Client canceled")
			return
		}

		select {
		case res := <-replyCh:
			for k, vs := range res.Headers {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(res.StatusCode)
			_, _ = w.Write(res.Body)
		case <-r.Context().Done():
			writePlainError(w, http.StatusRequestTimeout, "Gateway timeout waiting for pipeline")
		}
	}
}

// ===== Pipeline =====

func process(req RawRequestData) {
	start := time.Now()
	requestID := newRequestID()

	if !ipRateLimiter(req.IP) {
		req.ReplyCh <- normalizedError(requestID, http.StatusTooManyRequests, "RATE_LIMIT_IP", "Too Many Requests (IP)", time.Since(start))
		return
	}

	if !jwtChecker(req.Token) {
		req.ReplyCh <- normalizedError(requestID, http.StatusUnauthorized, "UNAUTHENTICATED", "Unauthorized (JWT)", time.Since(start))
		return
	}

	if !featureRateLimiter(req.Topic) {
		req.ReplyCh <- normalizedError(requestID, http.StatusTooManyRequests, "RATE_LIMIT_FEATURE", "Too Many Requests (Feature)", time.Since(start))
		return
	}

	res := routeToInternalService(req, requestID, start)
	req.ReplyCh <- res
}

// ===== Mock middlewares =====

func ipRateLimiter(ip string) bool         { return true }
func jwtChecker(token string) bool         { return strings.HasPrefix(token, "Bearer ") }
func featureRateLimiter(topic string) bool { return true }

// ===== Routing =====

func routeToInternalService(req RawRequestData, requestID string, start time.Time) GatewayResult {
	targetURL := ""
	switch req.Topic {
	case "auth/login", "auth/register":
		targetURL = "http://localhost:9090" + req.Path
	case "profile/get":
		targetURL = "http://localhost:9091" + req.Path
	case "profile/update":
		targetURL = "http://localhost:9092" + req.Path
	}

	ctx, cancel := context.WithTimeout(req.Ctx, 3*time.Second)
	defer cancel()

	ireq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, bytes.NewReader(req.Body))
	if err != nil {
		return normalizedError(requestID, http.StatusInternalServerError, "BUILD_REQUEST_FAILED", err.Error(), time.Since(start))
	}

	copySafeHeaders(req.Header, ireq.Header)
	ireq.Header.Set("X-Request-ID", requestID)
	ireq.Header.Set("X-Trace-ID", newRequestID())

	client := &http.Client{}
	resp, err := client.Do(ireq)
	latency := time.Since(start)
	if err != nil {
		return normalizedError(requestID, http.StatusBadGateway, "BAD_GATEWAY", "Internal service unreachable: "+err.Error(), latency)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	log.Println("====== Internal Service Response ======")
	log.Printf("RequestID: %s | Status: %d %s | Latency: %dms", requestID, resp.StatusCode, http.StatusText(resp.StatusCode), latency.Milliseconds())
	for k, v := range resp.Header {
		log.Printf("%s: %v", k, v)
	}
	log.Println("Body:", string(respBody))
	log.Println("======================================")

	out := map[string]interface{}{
		"request_id": requestID,
		"status":     "SUCCESS",
		"latency_ms": latency.Milliseconds(),
		"data":       json.RawMessage(respBody),
		"error":      nil,
	}

	status := http.StatusOK
	if resp.StatusCode >= 400 {
		out["status"] = "ERROR"
		out["error"] = map[string]interface{}{
			"upstream_status": resp.StatusCode,
			"message":         string(respBody),
		}
		out["data"] = nil
		status = http.StatusOK
	}

	body, _ := json.Marshal(out)
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("X-Gateway", "api-gateway")
	h.Set("X-Request-ID", requestID)

	return GatewayResult{
		StatusCode: status,
		Headers:    h,
		Body:       body,
	}
}

// ===== Helpers =====

func copySafeHeaders(src, dst http.Header) {
	if ct := src.Get("Content-Type"); ct != "" {
		dst.Set("Content-Type", ct)
	}
	if acc := src.Get("Accept"); acc != "" {
		dst.Set("Accept", acc)
	}
}

func normalizedError(requestID string, httpCode int, code string, message string, latency time.Duration) GatewayResult {
	payload := map[string]interface{}{
		"request_id": requestID,
		"status":     "ERROR",
		"latency_ms": latency.Milliseconds(),
		"data":       nil,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	body, _ := json.Marshal(payload)
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("X-Gateway", "api-gateway")
	h.Set("X-Request-ID", requestID)
	return GatewayResult{
		StatusCode: httpCode,
		Headers:    h,
		Body:       body,
	}
}

func writePlainError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(msg))
}

func newRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
