package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"backend/database"

	"github.com/stretchr/testify/assert"
)

func TestUpdateUser(t *testing.T) {

	router := setupTest()
	username := randomUsername()

	// 1️⃣ Register
	body := map[string]string{
		"username": username,
		"password": "1234",
	}

	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest(
		"POST",
		"/register",
		bytes.NewBuffer(jsonBody),
	)

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// debug
	if w.Code != 201 {
		t.Log("Register response:", w.Body.String())
	}

	assert.Equal(t, 201, w.Code)

	// 2️⃣ Get id from DB
	var id int

	err := database.DB.QueryRow(
		"SELECT id FROM users WHERE username=$1",
		username,
	).Scan(&id)

	assert.NoError(t, err)
	assert.NotZero(t, id)

	newUsername := randomUsername()

	updateBody := map[string]string{
		"username": newUsername,
		"password": "9999",
	}

	updateJSON, _ := json.Marshal(updateBody)

	req2, _ := http.NewRequest(
		"PUT",
		"/users/"+strconv.Itoa(id),
		bytes.NewBuffer(updateJSON),
	)

	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// debug
	if w2.Code != 200 {
		t.Log("Update response:", w2.Body.String())
	}

	assert.Equal(t, 200, w2.Code)

	// 4️⃣ Check DB
	var newName string

	err = database.DB.QueryRow(
		"SELECT username FROM users WHERE id=$1",
		id,
	).Scan(&newName)

	assert.NoError(t, err)
	assert.Equal(t, newUsername, newName)
}
