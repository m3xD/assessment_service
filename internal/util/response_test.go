package util

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseOK(t *testing.T) {
	rr := httptest.NewRecorder() // Dùng recorder làm ResponseWriter giả

	responseData := map[string]string{"message": "Success"}
	res := Response{
		StatusCode: http.StatusOK,
		Message:    "Operation successful",
		Data:       responseData,
	}

	ResponseOK(rr, res)

	// Kiểm tra Status Code
	assert.Equal(t, http.StatusOK, rr.Code, "Status code should be OK")

	// Kiểm tra Content-Type Header
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Content-Type should be application/json")

	// Kiểm tra Response Body
	var bodyResponse Response
	err := json.Unmarshal(rr.Body.Bytes(), &bodyResponse)
	require.NoError(t, err, "Failed to unmarshal response body")

	assert.Equal(t, res.StatusCode, bodyResponse.StatusCode)
	assert.Equal(t, res.Message, bodyResponse.Message)

	// So sánh Data (cần ép kiểu vì Data là interface{})
	bodyData, ok := bodyResponse.Data.(map[string]interface{})
	require.True(t, ok, "Response data should be a map")
	expectedData, ok := res.Data.(map[string]string)
	require.True(t, ok, "Expected data should be a map")

	// So sánh từng phần tử trong map data
	assert.Equal(t, expectedData["message"], bodyData["message"])

}

func TestResponseMap(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]interface{}{
		"status":  "CREATED",
		"id":      123,
		"isValid": true,
	}
	statusCode := http.StatusCreated

	ResponseMap(rr, data, statusCode)

	assert.Equal(t, statusCode, rr.Code, "Status code should match")
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var bodyResponse map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &bodyResponse)
	require.NoError(t, err)

	// So sánh map (JSON number sẽ là float64)
	assert.Equal(t, data["status"], bodyResponse["status"])
	assert.Equal(t, float64(data["id"].(int)), bodyResponse["id"]) // Ép kiểu float64
	assert.Equal(t, data["isValid"], bodyResponse["isValid"])
}

func TestResponseInterface(t *testing.T) {
	rr := httptest.NewRecorder()
	type CustomData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	data := CustomData{Name: "Test", Value: 42}
	statusCode := http.StatusOK

	ResponseInterface(rr, data, statusCode)

	assert.Equal(t, statusCode, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var bodyResponse CustomData
	err := json.Unmarshal(rr.Body.Bytes(), &bodyResponse)
	require.NoError(t, err)
	assert.Equal(t, data, bodyResponse)
}

func TestResponseError(t *testing.T) {
	rr := httptest.NewRecorder()
	res := Response{
		StatusCode: http.StatusUnauthorized,
		Message:    "Authentication required",
		Data:       nil, // Thường Data là nil cho lỗi
	}

	ResponseError(rr, res)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var bodyResponse Response
	err := json.Unmarshal(rr.Body.Bytes(), &bodyResponse)
	require.NoError(t, err)
	assert.Equal(t, res.StatusCode, bodyResponse.StatusCode)
	assert.Equal(t, res.Message, bodyResponse.Message)
	assert.Nil(t, bodyResponse.Data) // Kiểm tra Data là nil
}

func TestResponseError_WithData(t *testing.T) {
	rr := httptest.NewRecorder()
	errorData := map[string]string{"field": "email", "reason": "invalid format"}
	res := Response{
		StatusCode: http.StatusBadRequest,
		Message:    "Validation failed",
		Data:       errorData, // Data chứa chi tiết lỗi
	}

	ResponseError(rr, res)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var bodyResponse Response
	err := json.Unmarshal(rr.Body.Bytes(), &bodyResponse)
	require.NoError(t, err)
	assert.Equal(t, res.StatusCode, bodyResponse.StatusCode)
	assert.Equal(t, res.Message, bodyResponse.Message)

	// Kiểm tra Data lỗi
	bodyData, ok := bodyResponse.Data.(map[string]interface{})
	require.True(t, ok)
	expectedData := res.Data.(map[string]string)
	assert.Equal(t, expectedData["field"], bodyData["field"])
	assert.Equal(t, expectedData["reason"], bodyData["reason"])
}
