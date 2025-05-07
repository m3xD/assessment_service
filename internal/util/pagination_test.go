package util

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected PaginationParams
	}{
		{
			name: "Default values",
			url:  "/items",
			expected: PaginationParams{
				Page:    0,
				Limit:   10,
				Offset:  0,
				Search:  "",
				SortBy:  "created_at", // Default sort
				SortDir: "DESC",       // Default direction
				Filters: map[string]interface{}{},
			},
		},
		{
			name: "With page and pageSize",
			url:  "/items?page=2&pageSize=20",
			expected: PaginationParams{
				Page:    2,
				Limit:   20,
				Offset:  40, // 2 * 20
				Search:  "",
				SortBy:  "created_at",
				SortDir: "DESC",
				Filters: map[string]interface{}{},
			},
		},
		{
			name: "With search term",
			url:  "/items?search=keyword",
			expected: PaginationParams{
				Page:    0,
				Limit:   10,
				Offset:  0,
				Search:  "keyword",
				SortBy:  "created_at",
				SortDir: "DESC",
				Filters: map[string]interface{}{},
			},
		},
		{
			name: "With sort ascending",
			url:  "/items?sort=name,asc",
			expected: PaginationParams{
				Page:    0,
				Limit:   10,
				Offset:  0,
				Search:  "",
				SortBy:  "name", // Converted by toSnakeCase
				SortDir: "ASC",
				Filters: map[string]interface{}{},
			},
		},
		{
			name: "With sort descending (default dir)",
			url:  "/items?sort=email",
			expected: PaginationParams{
				Page:    0,
				Limit:   10,
				Offset:  0,
				Search:  "",
				SortBy:  "email",
				SortDir: "DESC",
				Filters: map[string]interface{}{},
			},
		},
		{
			name: "With sort camelCase",
			url:  "/items?sort=lastLogin,desc",
			expected: PaginationParams{
				Page:    0,
				Limit:   10,
				Offset:  0,
				Search:  "",
				SortBy:  "last_login", // Converted by toSnakeCase
				SortDir: "DESC",
				Filters: map[string]interface{}{},
			},
		},
		{
			name: "With invalid page/pageSize",
			url:  "/items?page=abc&pageSize=-5",
			expected: PaginationParams{
				Page:    0,  // Default page
				Limit:   10, // Default limit
				Offset:  0,
				Search:  "",
				SortBy:  "created_at",
				SortDir: "DESC",
				Filters: map[string]interface{}{},
			},
		},
		{
			name: "With multiple params",
			url:  "/items?page=1&pageSize=15&search=term&sort=updatedAt,asc",
			expected: PaginationParams{
				Page:    1,
				Limit:   15,
				Offset:  15, // 1 * 15
				Search:  "term",
				SortBy:  "updated_at", // Converted
				SortDir: "ASC",
				Filters: map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqURL, _ := url.Parse(tt.url)
			req := &http.Request{URL: reqURL}
			params := GetPaginationParams(req)
			// Compare field by field for clarity
			assert.Equal(t, tt.expected.Page, params.Page)
			assert.Equal(t, tt.expected.Limit, params.Limit)
			assert.Equal(t, tt.expected.Offset, params.Offset)
			assert.Equal(t, tt.expected.Search, params.Search)
			assert.Equal(t, tt.expected.SortBy, params.SortBy)
			assert.Equal(t, tt.expected.SortDir, params.SortDir)
			assert.Equal(t, tt.expected.Filters, params.Filters) // Filters are initialized but not populated by this func
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SimpleTest", "simple_test"},
		{"simpleTest", "simple_test"},
		{"already_snake", "already_snake"},
		{"UserID", "user_i_d"},
		{"ID", "i_d"},
		{"CreatedAt", "created_at"},
		{"A", "a"},
		{"", ""},
		{"simple", "simple"},
		{"URLShortener", "u_r_l_shortener"},
		{"userID", "user_i_d"},      // Test case from pagination sort
		{"createdAt", "created_at"}, // Test case from pagination sort
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, toSnakeCase(tt.input))
		})
	}
}

func TestCreatePaginationResponse(t *testing.T) {
	data := []map[string]string{
		{"id": "1", "name": "Item 1"},
		{"id": "2", "name": "Item 2"},
	}
	total := int64(25)
	params := PaginationParams{
		Page:    1,
		Limit:   10,
		Offset:  10,
		SortBy:  "name",
		SortDir: "ASC",
	}

	response := CreatePaginationResponse(data, total, params)

	respMap, ok := response.(map[string]interface{})
	require.True(t, ok, "Response should be a map")

	// Check basic pagination fields
	assert.Equal(t, (total), respMap["totalElements"]) // JSON numbers are float64
	assert.Equal(t, int64(3), respMap["totalPages"])   // ceil(25 / 10) = 3
	assert.Equal(t, (params.Page), respMap["number"])
	assert.Equal(t, (params.Limit), respMap["size"])
	assert.Equal(t, (len(data)), respMap["numberOfElements"])
	assert.False(t, respMap["first"].(bool))
	assert.False(t, respMap["last"].(bool)) // Page 1 is not the last page (page 2 is)
	assert.False(t, respMap["empty"].(bool))

	// Check content
	content, ok := respMap["content"].([]map[string]string)
	require.True(t, ok)
	assert.Equal(t, data, content)

	// Check pageable object
	pageable, ok := respMap["pageable"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, (params.Page), pageable["pageNumber"])
	assert.Equal(t, (params.Limit), pageable["pageSize"])
	assert.Equal(t, (params.Offset), pageable["offset"])
	assert.True(t, pageable["paged"].(bool))
	assert.False(t, pageable["unpaged"].(bool))

	// Check sort object (basic check as structure is simple)
	sortInfo, ok := respMap["sort"].(map[string]interface{})
	require.True(t, ok)
	assert.True(t, sortInfo["sorted"].(bool))
	assert.False(t, sortInfo["unsorted"].(bool))

	pageableSortInfo, ok := pageable["sort"].(map[string]interface{})
	require.True(t, ok)
	assert.True(t, pageableSortInfo["sorted"].(bool))
	assert.False(t, pageableSortInfo["unsorted"].(bool))

}

func TestCreatePaginationResponse_FirstPage(t *testing.T) {
	data := []map[string]string{{"id": "1"}}
	total := int64(5)
	params := PaginationParams{Page: 0, Limit: 2, Offset: 0}
	response := CreatePaginationResponse(data, total, params)
	respMap := response.(map[string]interface{})
	assert.True(t, respMap["first"].(bool))
	assert.False(t, respMap["last"].(bool))
}

func TestCreatePaginationResponse_LastPage(t *testing.T) {
	data := []map[string]string{{"id": "5"}}
	total := int64(5)
	params := PaginationParams{Page: 2, Limit: 2, Offset: 4} // Page 2 (0-indexed), limit 2
	response := CreatePaginationResponse(data, total, params)
	respMap := response.(map[string]interface{})
	assert.False(t, respMap["first"].(bool))
	assert.True(t, respMap["last"].(bool)) // TotalPages = ceil(5/2) = 3. Page 2 is last.
}

// TestBuildUserListQuery requires more complex setup or mocking of DB interactions,
// often better suited for integration tests or testing via service layer mocks.
// Basic tests could check parts of the generated query string.
func TestBuildUserListQuery_Basic(t *testing.T) {
	params := PaginationParams{
		Page:    1,
		Limit:   10,
		Offset:  10,
		Search:  "test",
		SortBy:  "name",
		SortDir: "ASC",
		Filters: map[string]interface{}{
			"role":   "student",
			"status": "active",
		},
	}

	query, countQuery, args, err := BuildUserListQuery(params)
	assert.NoError(t, err)

	// Check basic query structure
	assert.Contains(t, query, "SELECT id, full_name, email, role, status, phone, created_at, updated_at FROM users")
	assert.Contains(t, query, "WHERE (full_name ILIKE $1 OR email ILIKE $2) AND role = $3 AND status = $4")
	assert.Contains(t, query, "ORDER BY name ASC") // Assuming toSnakeCase handles 'name' correctly
	assert.Contains(t, query, "LIMIT $5 OFFSET $6")

	// Check count query structure
	assert.Contains(t, countQuery, "SELECT COUNT(*) FROM users")
	assert.Contains(t, countQuery, "WHERE (full_name ILIKE $1 OR email ILIKE $2) AND role = $3 AND status = $4")
	assert.NotContains(t, countQuery, "ORDER BY") // Count query shouldn't have ORDER BY
	assert.NotContains(t, countQuery, "LIMIT")    // Count query shouldn't have LIMIT/OFFSET

	// Check arguments
	assert.Len(t, args, 6)
	assert.Equal(t, "%test%", args[0])
	assert.Equal(t, "%test%", args[1])
	assert.Equal(t, "student", args[2])
	assert.Equal(t, "active", args[3])
	assert.Equal(t, 10, args[4]) // Limit
	assert.Equal(t, 10, args[5]) // Offset
}

func TestBuildUserListQuery_NoFilters(t *testing.T) {
	params := PaginationParams{
		Page:    0,
		Limit:   5,
		Offset:  0,
		SortBy:  "created_at", // Default snake_case
		SortDir: "DESC",
		Filters: map[string]interface{}{}, // No filters
	}

	query, countQuery, args, err := BuildUserListQuery(params)
	assert.NoError(t, err)

	assert.NotContains(t, query, "WHERE") // No WHERE clause expected
	assert.Contains(t, query, "ORDER BY created_at DESC")
	assert.Contains(t, query, "LIMIT $1 OFFSET $2")

	assert.NotContains(t, countQuery, "WHERE")

	assert.Len(t, args, 2)
	assert.Equal(t, 5, args[0]) // Limit
	assert.Equal(t, 0, args[1]) // Offset
}
