package middleware

import (
	"log"
	"net/http"
)

// CORSMiddleware thêm các header CORS vào response và xử lý preflight OPTIONS request.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ghi log để debug middleware (tuỳ chọn)
		log.Printf("CORSMiddleware: %s %s", r.Method, r.RequestURI)

		// Đặt header CORS
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Nếu là preflight request (OPTIONS), trả về ngay với status OK
		if r.Method == http.MethodOptions {
			log.Println("OPTIONS preflight request, trả về 200 OK")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
