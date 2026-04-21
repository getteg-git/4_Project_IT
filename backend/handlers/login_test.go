package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {

	setupTest()
	gin.SetMode(gin.TestMode)

	router := setupTest()

	body := map[string]string{
		"username": "john",
		"password": "1111",
	}

	jsonValue, _ := json.Marshal(body)

	req, _ := http.NewRequest(
		"POST",
		"/login",
		bytes.NewBuffer(jsonValue),
	)

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
