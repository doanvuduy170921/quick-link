package error

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHandlerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		err              error
		expectStatusCode int
		expectBodyErrMsg string
	}{
		{
			name:             "validation error",
			err:              NewValidationError("url must not be empty"),
			expectStatusCode: http.StatusBadRequest,
			expectBodyErrMsg: "url must not be empty",
		},
		{
			name:             "not found error",
			err:              NewNotFoundError("short code not found", errors.New("root cause")),
			expectStatusCode: http.StatusNotFound,
			expectBodyErrMsg: "short code not found",
		},
		{
			name:             "internal error - message hidden from client",
			err:              NewInternalError("create url error", errors.New("db connection failed")),
			expectStatusCode: http.StatusInternalServerError,
			expectBodyErrMsg: "internal server error",
		},
		{
			name: "unknown URLError type - falls to default",
			err: &URLError{
				Type:    ErrorType("some_future_type"),
				Message: "weird error",
			},
			expectStatusCode: http.StatusInternalServerError,
			expectBodyErrMsg: "internal server error",
		},
		{
			name:             "plain error - not *URLError at all",
			err:              errors.New("some random error"),
			expectStatusCode: http.StatusInternalServerError,
			expectBodyErrMsg: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			HandlerError(c, tt.err)

			require.Equal(t, tt.expectStatusCode, w.Code)
			require.Contains(t, w.Body.String(), tt.expectBodyErrMsg)

			if tt.expectStatusCode == http.StatusInternalServerError {
				require.NotContains(t, w.Body.String(), "db connection failed")
			}
		})
	}
}

func TestURLError_Error(t *testing.T) {
	err := NewValidationError("some message")
	require.Equal(t, "some message", err.Error())
}

func TestURLError_Unwrap(t *testing.T) {
	rootCause := errors.New("root cause")
	err := NewInternalError("wrapper message", rootCause)

	require.Equal(t, rootCause, err.Unwrap())
	require.True(t, errors.Is(err, rootCause))
}
