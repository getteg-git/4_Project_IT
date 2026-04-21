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

func TestDeleteUser(t *testing.T) {

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

	assert.Equal(t, 201, w.Code)

	// 2️⃣ ดึง id จาก DB ตรงๆ
	var id int

	err := database.DB.QueryRow(
		"SELECT id FROM users WHERE username=$1",
		body["username"],
	).Scan(&id)

	assert.NoError(t, err)
	assert.NotZero(t, id)

	// 3️⃣ Delete
	req2, _ := http.NewRequest(
		"DELETE",
		"/users/"+strconv.Itoa(id),
		nil,
	)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)

	// 4️⃣ เช็คว่าหายจริง
	var count int

	err = database.DB.QueryRow(
		"SELECT COUNT(*) FROM users WHERE id=$1",
		id,
	).Scan(&count)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
