package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Helper: Mock Next Handler (Tái sử dụng hoặc định nghĩa lại nếu cần) ---
type MockNextHandlerCors struct {
	Called bool
}

func (h *MockNextHandlerCors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Called = true
	w.WriteHeader(http.StatusOK) // Ghi status OK để biết handler đã chạy
}

func TestCORSMiddleware_RegularRequest(t *testing.T) {
	nextHandler := &MockNextHandlerCors{}
	middlewareChain := CORSMiddleware(nextHandler)

	req := httptest.NewRequest("GET", "/api/data", nil)
	rr := httptest.NewRecorder()

	middlewareChain.ServeHTTP(rr, req)

	// Kiểm tra Status Code (phải là OK từ next handler)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Kiểm tra CORS Headers đã được set
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Authorization, Content-Type", rr.Header().Get("Access-Control-Allow-Headers"))

	// Kiểm tra next handler đã được gọi
	assert.True(t, nextHandler.Called)
}

func TestCORSMiddleware_OptionsPreflightRequest(t *testing.T) {
	nextHandler := &MockNextHandlerCors{}
	middlewareChain := CORSMiddleware(nextHandler)

	req := httptest.NewRequest("OPTIONS", "/api/data", nil)
	// Thêm các header thường có trong preflight request (ví dụ)
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
	req.Header.Set("Origin", "http://example.com") // Origin header is crucial for CORS

	rr := httptest.NewRecorder()

	middlewareChain.ServeHTTP(rr, req)

	// Kiểm tra Status Code (phải là OK do middleware xử lý)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Kiểm tra CORS Headers đã được set
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Authorization, Content-Type", rr.Header().Get("Access-Control-Allow-Headers"))

	// Kiểm tra next handler KHÔNG được gọi
	assert.False(t, nextHandler.Called)
}
