package application_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
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
			expected:   `{"result":4}`,
		},
		{
			expression: `{"expression":"2+2*2"}`,
			expected:   `{"result":6}`,
		},
		{
			expression: `{"expression":"(2+2)*2"}`,
			expected:   `{"result":8}`,
		},
		{
			expression: `{"expression":"20+20/0.5*(3+2)"}`,
			expected:   `{"result":220}`,
		},
		{
			expression: `{"expression":"1/2"}`,
			expected:   `{"result":0.5}`,
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
			if strings.Join(strings.Fields(w.Body.String()), "") != testCase.expected {
				t.Errorf("expected %s but got %s", testCase.expected, w.Body.String())
			}
		})
	}
	testCasesInvalid := []struct {
		expression    string
		method        string
		expectedError string
		expectedCode  int
	}{
		{
			`{"expression":"2+2*"}`,
			"POST",
			"expression is not valid",
			http.StatusUnprocessableEntity,
		},
		{
			`{"expression":""}`,
			"POST",
			"expression is empty",
			http.StatusUnprocessableEntity,
		},
		{
			`{"expression":"2+2*"}`,
			"GET",
			"not POST method",
			http.StatusMethodNotAllowed,
		},
		{
			`{"expression":"2+2}`,
			"POST",
			"invalid JSON format",
			http.StatusBadRequest,
		},
		{
			`{"expression":"1/0"}`,
			"POST",
			"division by zero",
			http.StatusUnprocessableEntity,
		},
	}
	for _, testCase := range testCasesInvalid {
		t.Run(testCase.expression, func(t *testing.T) {
			req := httptest.NewRequest(testCase.method, "/", bytes.NewBufferString(testCase.expression))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			calcHandler := http.HandlerFunc(application.CalcHandler)
			calcHandler.ServeHTTP(w, req)
			if body := w.Body.String(); !bytes.Contains([]byte(body), []byte(testCase.expectedError)) {
				t.Errorf("expected error %s but got %s", testCase.expectedError, w.Body.String())
			}
			if w.Code != testCase.expectedCode {
				t.Errorf("expected code %d but got %d", testCase.expectedCode, w.Code)
			}
		})
	}
}
