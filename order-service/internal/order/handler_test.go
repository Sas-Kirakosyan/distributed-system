package order

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newTestRouter builds the real router with a repository that has no pool.
// The tests below only hit the input checks, which return 400 before any
// DB call, so no database is needed. Paths that reach the DB are covered
// by scripts/smoke-test.sh against a real Postgres.
func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(NewRepository(nil)).Register(r)
	return r
}

func TestCreate_BadInput(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"not JSON", `hello`},
		{"empty body", ``},
		{"missing user_id", `{"amount": 1999}`},
		{"missing amount", `{"user_id": "u1"}`},
		{"amount is zero", `{"user_id": "u1", "amount": 0}`},
		{"amount below zero", `{"user_id": "u1", "amount": -5}`},
		{"amount is not a whole number", `{"user_id": "u1", "amount": 19.99}`},
		{"user_id too long", `{"user_id": "` + strings.Repeat("a", 65) + `", "amount": 1999}`},
	}

	router := newTestRouter()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, http.StatusBadRequest, w.Body.String())
			}
		})
	}
}

func TestGet_InvalidID(t *testing.T) {
	router := newTestRouter()

	for _, id := range []string{"abc", "123", "not-a-uuid"} {
		t.Run(id, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/orders/"+id, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}
