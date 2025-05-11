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
	"github.com/MoodyShoo/go-http-calculator/internal/orchestrator"
)

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
			w := httptest.NewRecorder()

			db, _ := database.NewInMemoryDatabase()
			orchestrator.New(db).CalculateHandler(w, req)

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

			for _, expr := range tc.expressions {
				req := httptest.NewRequest(http.MethodPost, orchestrator.CalculateRoute, bytes.NewBufferString(expr))
				w := httptest.NewRecorder()
				o.CalculateHandler(w, req)

				if w.Code != http.StatusAccepted {
					t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
				}
			}

			req := httptest.NewRequest(http.MethodGet, orchestrator.ExpressionsRoute, nil)
			w := httptest.NewRecorder()
			o.ExpressionsHandler(w, req)

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

			if tc.expression != "" {
				req := httptest.NewRequest(http.MethodPost, orchestrator.CalculateRoute, bytes.NewBufferString(tc.expression))
				w := httptest.NewRecorder()
				o.CalculateHandler(w, req)

				if w.Code != http.StatusAccepted {
					t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
				}
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s%d", orchestrator.ExpressionIdRoute, tc.id), nil)
			w := httptest.NewRecorder()
			o.ExpressionIdHandler(w, req)

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
