package util

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type PaginationParams struct {
	Page    int
	Limit   int
	Offset  int
	Search  string
	SortBy  string
	SortDir string
	Filters map[string]interface{}
}

func GetPaginationParams(r *http.Request) PaginationParams {
	page := 0
	limit := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val >= 0 {
			page = val
		}
	}

	if pageStr := r.URL.Query().Get("pageSize"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val >= 0 {
			limit = val
		}
	}

	// Calculate offset
	offset := page * limit

	// Get search term
	search := r.URL.Query().Get("search")

	// Get sort parameters
	sortParam := r.URL.Query().Get("sort")
	sortBy := "created_at"
	sortDir := "DESC"

	if sortParam != "" {
		sortParts := strings.Split(sortParam, ",")
		if len(sortParts) > 0 && sortParts[0] != "" {
			sortBy = toSnakeCase(sortParts[0])
		}

		if len(sortParts) > 1 {
			if strings.EqualFold(sortParts[1], "asc") {
				sortDir = "ASC"
			}
		}
	}

	return PaginationParams{
		Page:    page,
		Limit:   limit,
		Offset:  offset,
		Search:  search,
		SortBy:  sortBy,
		SortDir: sortDir,
		Filters: make(map[string]interface{}),
	}
}

func CreatePaginationResponse(data interface{}, total int64, params PaginationParams) interface{} {
	totalPages := (total + int64(params.Limit) - 1) / int64(params.Limit)

	return map[string]interface{}{
		"content":          data,
		"totalElements":    total,
		"totalPages":       totalPages,
		"number":           params.Page,
		"size":             params.Limit,
		"numberOfElements": reflect.ValueOf(data).Len(),
		"first":            params.Page == 0,
		"last":             int64(params.Page) == totalPages-1,
		"empty":            reflect.ValueOf(data).Len() == 0,
		"pageable": map[string]interface{}{
			"pageNumber": params.Page,
			"pageSize":   params.Limit,
			"offset":     params.Offset,
			"paged":      true,
			"unpaged":    false,
			"sort": map[string]interface{}{
				"sorted":   true,
				"unsorted": false,
			},
		},
		"sort": map[string]interface{}{
			"sorted":   true,
			"unsorted": false,
		},
		"activeUsers":   0,
		"inactiveUsers": 0,
	}
}

// Helper to convert camelCase to snake_case for DB column names
func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// BuildUserListQuery constructs SQL query for user listing with dynamic filters
func BuildUserListQuery(params PaginationParams) (string, string, []interface{}, error) {
	baseQuery := "SELECT id, full_name, email, role, status, phone, created_at, updated_at FROM users"
	countQuery := "SELECT COUNT(*) FROM users"

	whereConditions := []string{}
	args := []interface{}{}
	argPosition := 1

	// Apply search filter
	if params.Search != "" {
		whereConditions = append(whereConditions,
			fmt.Sprintf("(full_name ILIKE $%d OR email ILIKE $%d)", argPosition, argPosition+1))
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argPosition += 2
	}

	// Apply role filter
	if role, ok := params.Filters["role"].(string); ok && role != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("role = $%d", argPosition))
		args = append(args, role)
		argPosition++
	}

	// Apply status filter
	if status, ok := params.Filters["status"].(string); ok && status != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argPosition))
		args = append(args, status)
		argPosition++
	}

	// Combine where clauses
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Apply sorting
	sortClause := " ORDER BY created_at DESC"
	if params.SortBy != "" {
		sortClause = fmt.Sprintf(" ORDER BY %s %s", params.SortBy, params.SortDir)
	}

	// Add pagination to the query
	limitOffsetClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
	args = append(args, params.Limit, params.Offset)

	// Construct final queries
	fullQuery := baseQuery + whereClause + sortClause + limitOffsetClause
	countQueryWithWhere := countQuery + whereClause

	return fullQuery, countQueryWithWhere, args, nil
}
