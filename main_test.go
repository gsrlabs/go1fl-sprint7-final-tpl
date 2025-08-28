package main

import (
	"fmt"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		city  string
		count int
		want  int
	}{
		{"moscow", 0, 0},
		{"moscow", 1, 1},
		{"moscow", 2, 2},
		{"moscow", 100, min(100, len(cafeList["moscow"]))},

		{"tula", 0, 0},
		{"tula", 1, 1},
		{"tula", 2, 2},
		{"tula", 100, min(100, len(cafeList["tula"]))},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		requestURL := fmt.Sprintf("/cafe?count=%d&city=%s", v.count, v.city)
		req := httptest.NewRequest("GET", requestURL, nil)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := splitResponseBody(response.Body.String())

		assert.Len(t, body, v.want)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		city      string
		search    string
		wantCount int
	}{
		{"moscow", "фасоль", 0},
		{"moscow", "кофе", 2},
		{"moscow", "вилка", 1},
		{"moscow", "", 5},

		{"tula", "фасоль", 0},
		{"tula", "мир", 1},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		requestURL := fmt.Sprintf("/cafe?city=%s&search=%s", v.city, v.search)
		req := httptest.NewRequest("GET", requestURL, nil)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := splitResponseBody(response.Body.String())

		assert.Len(t, body, v.wantCount)
		for _, str := range body {
			assert.Contains(t, strings.ToLower(str), strings.ToLower(v.search))
		}
	}
}

func splitResponseBody(body string) []string {
	if body == "" {
		return []string{}
	}
	return strings.Split(body, ",")
}
