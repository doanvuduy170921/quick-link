package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/doanvuduy170921/quick-link/internal/mocks"
	urlErr "github.com/doanvuduy170921/quick-link/internal/url/error"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestURLHandler_ShortenURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name             string
		requestBody      ShortenInput
		rawBody          string
		mockSetUp        func(*mocks.MockUseCase)
		ExpectStatusCode int
	}{
		{
			name: "success",
			requestBody: ShortenInput{
				URL: "http://google.com",
			},
			mockSetUp: func(m *mocks.MockUseCase) {
				m.EXPECT().GenerateKey(mock.Anything, "http://google.com").Return("abcd123", nil).Once()
			},
			ExpectStatusCode: http.StatusOK,
		},
		{
			name: "malformed json syntax error",
			requestBody: ShortenInput{
				URL: "http://google.com",
			},
			mockSetUp:        func(m *mocks.MockUseCase) {},
			rawBody:          `{"url": "http://google.com"`,
			ExpectStatusCode: http.StatusBadRequest,
		},
		{
			name: " body invalid",
			requestBody: ShortenInput{
				URL: "abcd123",
			},
			mockSetUp: func(m *mocks.MockUseCase) {
			},
			ExpectStatusCode: http.StatusBadRequest,
		},
		{
			name: " body empty",
			requestBody: ShortenInput{
				URL: "",
			},
			mockSetUp: func(m *mocks.MockUseCase) {

			},
			ExpectStatusCode: http.StatusBadRequest,
		},
		{
			name:        "usecase returns validation error",
			requestBody: ShortenInput{URL: "http://valid-format-but-empty-in-usecase.com"},
			mockSetUp: func(m *mocks.MockUseCase) {
				m.EXPECT().
					GenerateKey(mock.Anything, mock.Anything).
					Return("", urlErr.NewValidationError("url must not be empty")).
					Once()
			},
			ExpectStatusCode: http.StatusBadRequest,
		},
		{
			name:        "usecase returns internal error",
			requestBody: ShortenInput{URL: "http://valid.com"},
			mockSetUp: func(m *mocks.MockUseCase) {
				m.EXPECT().
					GenerateKey(mock.Anything, mock.Anything).
					Return("", urlErr.NewInternalError("create url error", errors.New("db connection failed"))).
					Once()
			},
			ExpectStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := mocks.NewMockUseCase(t)
			tc.mockSetUp(mockUC)

			s := NewURLHandler(mockUC)
			router := gin.New()
			router.POST("/shorten", s.ShortenURL)

			var body []byte
			if tc.rawBody != "" {
				body = []byte(tc.rawBody)
			} else {
				bodyByte, err := json.Marshal(tc.requestBody)
				require.NoError(t, err)
				body = bodyByte
			}

			req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, tc.ExpectStatusCode, w.Code)
			if tc.name == "usecase returns internal error" {
				require.NotContains(t, w.Body.String(), "db connection failed")
			}

			if tc.name == "success" {
				var resp map[string]string
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				require.Equal(t, "abcd123", resp["code"])
			}
		})
	}

}

func TestURLHandler_Redirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		code             string
		mockSetUp        func(*mocks.MockUseCase)
		ExpectStatusCode int
	}{
		{
			name: "success",
			code: "abcd123",
			mockSetUp: func(m *mocks.MockUseCase) {
				m.EXPECT().Redirect(mock.Anything, "abcd123").Return("http://google.com", nil).Once()
			},
			ExpectStatusCode: http.StatusFound,
		},
		{
			name: "not found",
			code: "notexist",
			mockSetUp: func(m *mocks.MockUseCase) {
				m.EXPECT().Redirect(mock.Anything, "notexist").
					Return("", urlErr.NewNotFoundError("short code not found", nil)).Once()
			},
			ExpectStatusCode: http.StatusNotFound,
		},
		{
			name: "internal error",
			code: "abcd123",
			mockSetUp: func(m *mocks.MockUseCase) {
				m.EXPECT().Redirect(mock.Anything, "abcd123").
					Return("", urlErr.NewInternalError("get url error", errors.New("db down"))).Once()
			},
			ExpectStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := mocks.NewMockUseCase(t)
			tc.mockSetUp(mockUC)
			s := NewURLHandler(mockUC)
			router := gin.New()
			router.GET("/redirect/:code", s.Redirect)

			req := httptest.NewRequest(http.MethodGet, "/redirect/"+tc.code, nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, tc.ExpectStatusCode, w.Code)

			if tc.name == "success" {
				require.Equal(t, "http://google.com", w.Header().Get("Location"))
			}

		})
	}
}
