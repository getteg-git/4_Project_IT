package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRepairListAndUploadsRequireStaffSession(t *testing.T) {
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	router := gin.New()
	SetupRoutes(router)

	for _, path := range []string{"/api/repairs", "/uploads/example.jpg"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("GET %s status = %d, want %d", path, recorder.Code, http.StatusUnauthorized)
		}
	}
}
