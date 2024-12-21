package application_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pAran0k/calc_go/application"
)

func TestCases(t *testing.T) {
	testCasesSuccess := []struct {
		expression string
		expected   string
	}{
		{
			expression: `{"expression":"2+2"}`,
			expected:   `{"result":"4"}`,
		},
		{
			expression: `{"expression":"2+2*2"}`,
			expected:   `{"result":"6"}`,
		},
		{
			expression: `{"expression":"(2+2)*2"}`,
			expected:   `{"result":"8"}`,
		},
		{
			expression: `{"expression":"20+20/0.5*(3+2)"}`,
			expected:   `{"result":"220"}`,
		},
		{
			expression: `{"expression":"1/2"}`,
			expected:   `{"result":"0.5"}`,
		},
	}
	for _, testCase := range testCasesSuccess {
		t.Run(testCase.expression, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/",
				bytes.NewBufferString(testCase.expression))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			calcHandler := http.HandlerFunc(application.CalcHandler)
			calcHandler.ServeHTTP(w, req)
			if w.Body.String() != testCase.expected {
				t.Errorf("expected %s but got %s", testCase.expected, w.Body.String())
			}
		})
	}
}
