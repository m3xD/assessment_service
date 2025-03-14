package middleware

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type LogMiddleware struct {
	log *zap.Logger
}

func NewLogMiddleware(log *zap.Logger) *LogMiddleware {
	return &LogMiddleware{log: log}
}

func (l *LogMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Lấy thời gian bắt đầu xử lý request
		start := time.Now()

		// Log thông tin request ban đầu
		l.log.Info(fmt.Sprintf("Started %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr))

		// Chuyển request sang middleware/handler tiếp theo
		next.ServeHTTP(w, r)

		// Sau khi xử lý xong, log thời gian hoàn thành request
		duration := time.Since(start)
		l.log.Info(fmt.Sprintf("Completed %s in %v", r.RequestURI, duration))
	})
}
