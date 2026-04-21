package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchUsers(t *testing.T) {

	setupTest()

	router := setupTest()

	req, _ := http.NewRequest(
		"GET",
		"/users/search?q=john",
		nil,
	)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
