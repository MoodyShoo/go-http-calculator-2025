package orchestrator_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/MoodyShoo/go-http-calculator/internal/database"
	"github.com/MoodyShoo/go-http-calculator/internal/middleware"
	"github.com/MoodyShoo/go-http-calculator/internal/orchestrator"
)

func registerAndLogin(t *testing.T, o *orchestrator.Orchestrator) string {
	registerReq := httptest.NewRequest(http.MethodPost, orchestrator.RegisterRoute, bytes.NewBufferString(`{"login":"test","password":"1234"}`))
	registerW := httptest.NewRecorder()
	o.RegisterHandler(registerW, registerReq)
	if registerW.Code != http.StatusOK {
		t.Fatalf("register failed: status = %d, body = %s", registerW.Code, registerW.Body.String())
	}

	loginReq := httptest.NewRequest(http.MethodPost, orchestrator.LoginRoute, bytes.NewBufferString(`{"login":"test","password":"1234"}`))
	loginW := httptest.NewRecorder()
	o.LoginHandler(loginW, loginReq)
	if loginW.Code != http.StatusOK {
		t.Fatalf("login failed: status = %d, body = %s", loginW.Code, loginW.Body.String())
	}

	var resp struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(loginW.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	return resp.Token
}

func TestCalculateRoute(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		request    string
		want       string
	}{
		{
			name:       "Valid expression",
			statusCode: http.StatusAccepted,
			request:    `{"expression": "2+2"}`,
			want:       `{"id":1}`,
		},
		{
			name:       "Invalid JSON",
			statusCode: http.StatusUnprocessableEntity,
			request:    `{"expression": 2+2}`,
			want:       `{"error":"unprocessable entity"}`,
		},
		{
			name:       "Invalid expression",
			statusCode: http.StatusInternalServerError,
			request:    `{"expression": "2+2-"}`,
			want:       `{"error":"failed to create tasks: not enough operands for operator: -"}`,
		},
		{
			name:       "Empty expression",
			statusCode: http.StatusUnprocessableEntity,
			request:    `{"expression": ""}`,
			want:       `{"error":"unprocessable entity"}`,
		},
		{
			name:       "Malformed JSON",
			statusCode: http.StatusUnprocessableEntity,
			request:    `{"expression": "2+2"`,
			want:       `{"error":"unprocessable entity"}`,
		},
		{
			name:       "Missing expression field",
			statusCode: http.StatusUnprocessableEntity,
			request:    `{}`,
			want:       `{"error":"unprocessable entity"}`,
		},
		{
			name:       "Large expression",
			statusCode: http.StatusAccepted,
			request:    `{"expression": "2+2*3-4/2+6*5-10+8"}`,
			want:       `{"id":1}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, orchestrator.CalculateRoute, bytes.NewReader([]byte(tc.request)))

			db, _ := database.NewInMemoryDatabase()
			o := orchestrator.New(db)
			token := registerAndLogin(t, o)

			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			handler := middleware.AuthMiddleware(&o.Ts, o.CalculateHandler)
			handler.ServeHTTP(w, req)

			if status := w.Code; status != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, status)
			}

			if got := w.Body.String(); got != tc.want {
				t.Errorf("Expected body %s, got %s", tc.want, got)
			}
		})
	}
}

func TestExpressionsHandler(t *testing.T) {
	cases := []struct {
		name        string
		expressions []string
		statusCode  int
		want        string
	}{
		{
			name:        "One valid expression",
			expressions: []string{`{"expression": "2+2"}`},
			statusCode:  http.StatusOK,
			want:        `{"expressions":[{"id":1,"expression":"2+2","status":"pending","result":0}]}`,
		},
		{
			name:        "Multiple valid expressions",
			expressions: []string{`{"expression": "2+2"}`, `{"expression": "3*3"}`},
			statusCode:  http.StatusOK,
			want:        `{"expressions":[{"id":1,"expression":"2+2","status":"pending","result":0},{"id":2,"expression":"3*3","status":"pending","result":0}]}`,
		},
		{
			name:        "Empty list",
			expressions: []string{},
			statusCode:  http.StatusOK,
			want:        `{"expressions":[]}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, _ := database.NewInMemoryDatabase()
			o := orchestrator.New(db)

			token := registerAndLogin(t, o)

			for _, expr := range tc.expressions {
				req := httptest.NewRequest(http.MethodPost, orchestrator.CalculateRoute, bytes.NewBufferString(expr))
				req.Header.Set("Authorization", "Bearer "+token)

				handler := middleware.AuthMiddleware(&o.Ts, o.CalculateHandler)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				if w.Code != http.StatusAccepted {
					t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
				}
			}

			req := httptest.NewRequest(http.MethodGet, orchestrator.ExpressionsRoute, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			handler := middleware.AuthMiddleware(&o.Ts, o.ExpressionsHandler)
			handler.ServeHTTP(w, req)

			if status := w.Code; status != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, status)
			}

			var got map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			var want map[string]interface{}
			if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
				t.Fatalf("Failed to unmarshal expected body: %v", err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("Expected body %v, got %v", want, got)
			}
		})
	}
}

func TestExpressionIdHandler(t *testing.T) {
	cases := []struct {
		name       string
		expression string
		id         int
		statusCode int
		want       string
	}{
		{
			name:       "Valid expression ID",
			expression: `{"expression": "2+2"}`,
			id:         1,
			statusCode: http.StatusOK,
			want:       `{"id":1,"expression":"2+2","status":"pending","result":0}`,
		},
		{
			name:       "Invalid expression ID",
			expression: `{"expression": "2+2"}`,
			id:         999,
			statusCode: http.StatusNotFound,
			want:       `{"error":"expression not found"}`,
		},
		{
			name:       "Invalid ID format",
			expression: `{"expression": "2+2"}`,
			id:         -1,
			statusCode: http.StatusNotFound,
			want:       `{"error":"expression not found"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, _ := database.NewInMemoryDatabase()
			o := orchestrator.New(db)
			token := registerAndLogin(t, o)

			if tc.expression != "" {
				req := httptest.NewRequest(http.MethodPost, orchestrator.CalculateRoute, bytes.NewBufferString(tc.expression))
				req.Header.Set("Authorization", "Bearer "+token)

				handler := middleware.AuthMiddleware(&o.Ts, o.CalculateHandler)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				if w.Code != http.StatusAccepted {
					t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
				}
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s%d", orchestrator.ExpressionIdRoute, tc.id), nil)
			req.Header.Set("Authorization", "Bearer "+token)

			handler := middleware.AuthMiddleware(&o.Ts, o.ExpressionIdHandler)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if status := w.Code; status != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, status)
			}

			var got map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			var want map[string]interface{}
			if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
				t.Fatalf("Failed to unmarshal expected body: %v", err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("Expected body %v, got %v", want, got)
			}
		})
	}
}
