package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {

	setupTest()

	gin.SetMode(gin.TestMode)

	router := setupTest()

	username := "testuser_" + time.Now().Format("20060102150405")

	body := map[string]string{
		"username": username,
		"password": "1234",
	}

	jsonValue, _ := json.Marshal(body)

	req, _ := http.NewRequest(
		"POST",
		"/register",
		bytes.NewBuffer(jsonValue),
	)

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Log("Response:", w.Body.String())
	}

	assert.Equal(t, 201, w.Code)
}
