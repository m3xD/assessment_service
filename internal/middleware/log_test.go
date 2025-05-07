package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest" // Sử dụng zaptest để tạo logger đơn giản
)

// --- Helper: Mock Next Handler (Tái sử dụng) ---
type MockNextHandlerLog struct {
	Called bool
}

func (h *MockNextHandlerLog) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Called = true
	w.WriteHeader(http.StatusNoContent) // Trả về status khác để phân biệt
}

func TestLoggingMiddleware(t *testing.T) {
	logger := zaptest.NewLogger(t) // Tạo logger test (output sẽ bị discard hoặc hiển thị tùy cấu hình test)
	logMiddleware := NewLogMiddleware(logger)
	nextHandler := &MockNextHandlerLog{}

	middlewareChain := logMiddleware.LoggingMiddleware(nextHandler)

	req := httptest.NewRequest("GET", "/test/path", nil)
	rr := httptest.NewRecorder()

	middlewareChain.ServeHTTP(rr, req)

	// Kiểm tra Status Code (phải là NoContent từ next handler)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Kiểm tra next handler đã được gọi
	assert.True(t, nextHandler.Called)

	// Lưu ý: Việc assert log output cụ thể ở đây khá phức tạp.
	// Test này chủ yếu đảm bảo middleware chạy và gọi next handler.
	// Bạn có thể cần các kỹ thuật nâng cao hơn (custom zap core) nếu muốn kiểm tra log chi tiết.
}
