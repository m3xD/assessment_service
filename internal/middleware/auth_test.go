package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time" // Cần thiết cho JWT claims

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mock Jwt Service ---
type MockJwtService struct {
	mock.Mock
}

func (m *MockJwtService) GenerateToken(userID string, role string) (string, error) {
	args := m.Called(userID, role)
	return args.String(0), args.Error(1)
}

func (m *MockJwtService) ValidateToken(token string) (jwt.MapClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Handle potential nil map but non-nil interface{}
	claims, ok := args.Get(0).(jwt.MapClaims)
	if !ok && args.Get(0) != nil {
		// If it's not nil but not MapClaims, return an error or panic
		// depending on how strict you want the mock to be.
		// Returning nil, nil might be acceptable if the error case is handled separately.
		return nil, errors.New("mock ValidateToken returned non-nil value of incorrect type")
	}
	return claims, args.Error(1)
}

// --- Helper: Mock Next Handler ---
type MockNextHandler struct {
	Called bool
	Ctx    context.Context // Để kiểm tra context được truyền vào
}

func (h *MockNextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Called = true
	h.Ctx = r.Context()          // Lưu lại context
	w.WriteHeader(http.StatusOK) // Ghi status OK để biết handler đã chạy
}

// --- Test AuthMiddleware ---

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	validToken := "valid.jwt.token"
	expectedClaims := jwt.MapClaims{
		"userID": "user123",
		"role":   "student",
		"exp":    float64(time.Now().Add(time.Hour).Unix()), // Phải là float64
	}

	// Expect ValidateToken to be called with the token from header
	mockJwt.On("ValidateToken", validToken).Return(expectedClaims, nil)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.AuthMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rr.Code, "Next handler should be called and return OK")
	assert.True(t, nextHandler.Called, "Next handler should be called")
	mockJwt.AssertExpectations(t)

	// Kiểm tra context
	require.NotNil(t, nextHandler.Ctx, "Context should not be nil in next handler")
	userClaims, ok := nextHandler.Ctx.Value("user").(jwt.MapClaims)
	require.True(t, ok, "User claims should be in context")
	assert.Equal(t, expectedClaims["userID"], userClaims["userID"])
	assert.Equal(t, expectedClaims["role"], userClaims["role"])
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	req := httptest.NewRequest("GET", "/protected", nil) // No Authorization header
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.AuthMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, nextHandler.Called, "Next handler should not be called")
	mockJwt.AssertNotCalled(t, "ValidateToken", mock.Anything) // ValidateToken không được gọi
}

func TestAuthMiddleware_InvalidHeaderFormat_NoBearer(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat") // Thiếu "Bearer "
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.AuthMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, nextHandler.Called)
	mockJwt.AssertNotCalled(t, "ValidateToken", mock.Anything)
}

func TestAuthMiddleware_InvalidHeaderFormat_OnlyBearer(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer") // Chỉ có "Bearer"
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.AuthMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, nextHandler.Called)
	mockJwt.AssertNotCalled(t, "ValidateToken", mock.Anything)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	invalidToken := "invalid.jwt.token"
	validationError := errors.New("token validation failed")

	// Expect ValidateToken to be called and return an error
	mockJwt.On("ValidateToken", invalidToken).Return(nil, validationError)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.AuthMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, nextHandler.Called)
	mockJwt.AssertExpectations(t)
}

// --- Test ACLMiddleware ---

func TestACLMiddleware_AllowedRole(t *testing.T) {
	mockJwt := new(MockJwtService) // Không cần mock Jwt ở đây vì context đã được chuẩn bị
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	allowedRoles := []string{"admin", "teacher"}
	userClaims := jwt.MapClaims{"role": "teacher"} // Role được phép

	req := httptest.NewRequest("GET", "/admin/resource", nil)
	// Tạo context với claims giả lập (như AuthMiddleware đã làm)
	ctx := context.WithValue(req.Context(), "user", userClaims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.ACLMiddleware(allowedRoles...)(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Next handler should be called for allowed role")
	assert.True(t, nextHandler.Called)
}

func TestACLMiddleware_ForbiddenRole(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	allowedRoles := []string{"admin", "teacher"}
	userClaims := jwt.MapClaims{"role": "student"} // Role không được phép

	req := httptest.NewRequest("GET", "/admin/resource", nil)
	ctx := context.WithValue(req.Context(), "user", userClaims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.ACLMiddleware(allowedRoles...)(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.False(t, nextHandler.Called, "Next handler should not be called for forbidden role")
}

func TestACLMiddleware_NoUserInContext(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	allowedRoles := []string{"admin"}

	req := httptest.NewRequest("GET", "/admin/resource", nil) // Context không có key "user"
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.ACLMiddleware(allowedRoles...)(nextHandler)

	// Middleware sẽ panic nếu context không có "user" và không được ép kiểu jwt.MapClaims
	// Chúng ta cần bắt panic này hoặc đảm bảo AuthMiddleware luôn chạy trước.
	// Trong thực tế, nếu AuthMiddleware chạy trước, trường hợp này sẽ không xảy ra.
	// Để test an toàn, ta có thể kiểm tra panic.
	assert.Panics(t, func() {
		middlewareChain.ServeHTTP(rr, req)
	}, "ACLMiddleware should panic if 'user' claim is missing or not MapClaims")

	// Hoặc, nếu bạn muốn nó trả về lỗi thay vì panic:
	// middlewareChain.ServeHTTP(rr, req)
	// assert.Equal(t, http.StatusInternalServerError, rr.Code) // Hoặc Forbidden
	// assert.False(t, nextHandler.Called)
}

// --- Test OwnerMiddleware ---

func TestOwnerMiddleware_AdminRole(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	userClaims := jwt.MapClaims{"role": "admin", "userID": "admin1"}

	req := httptest.NewRequest("GET", "/users/otherUser/data", nil) // Admin truy cập tài nguyên của người khác
	ctx := context.WithValue(req.Context(), "user", userClaims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.OwnerMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, nextHandler.Called)
}

func TestOwnerMiddleware_UserAccessOwnData_WithMatchingID(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	userID := "user123"
	userClaims := jwt.MapClaims{"role": "user", "userID": userID}

	req := httptest.NewRequest("GET", "/users/data?id="+userID, nil) // User truy cập data với id khớp
	ctx := context.WithValue(req.Context(), "user", userClaims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.OwnerMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, nextHandler.Called)
}

func TestOwnerMiddleware_UserAccessOwnData_WithoutIDParam(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	userID := "user123"
	userClaims := jwt.MapClaims{"role": "user", "userID": userID}

	req := httptest.NewRequest("GET", "/users/mydata", nil) // User truy cập endpoint không yêu cầu ID param
	ctx := context.WithValue(req.Context(), "user", userClaims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.OwnerMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, nextHandler.Called)
}

func TestOwnerMiddleware_UserAccessOtherData_Forbidden(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	userID := "user123"
	otherUserID := "user456"
	userClaims := jwt.MapClaims{"role": "user", "userID": userID}

	req := httptest.NewRequest("GET", "/users/otherdata?id="+otherUserID, nil) // User cố truy cập data của người khác
	ctx := context.WithValue(req.Context(), "user", userClaims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.OwnerMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.False(t, nextHandler.Called)
}

func TestOwnerMiddleware_OtherRole_Forbidden(t *testing.T) {
	mockJwt := new(MockJwtService)
	authMiddleware := NewAuthMiddleware(mockJwt)
	nextHandler := &MockNextHandler{}

	userClaims := jwt.MapClaims{"role": "guest", "userID": "guest1"} // Role không phải admin/user

	req := httptest.NewRequest("GET", "/users/data?id=guest1", nil)
	ctx := context.WithValue(req.Context(), "user", userClaims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	middlewareChain := authMiddleware.OwnerMiddleware()(nextHandler)
	middlewareChain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.False(t, nextHandler.Called)
}
