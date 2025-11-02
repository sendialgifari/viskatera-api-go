package models

import (
	"time"
)

// Standard API Response Structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Success Response
func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// Error Response
func ErrorResponse(message string, code string, details string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

// Paginated Response
func PaginatedResponse(message string, data interface{}, page, perPage, total int) APIResponse {
	totalPages := (total + perPage - 1) / perPage
	return APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      &Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages},
		Timestamp: time.Now(),
	}
}
