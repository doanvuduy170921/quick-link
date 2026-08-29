//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	handlerhttp "github.com/doanvuduy170921/quick-link/internal/url/delivery/http"
	"github.com/doanvuduy170921/quick-link/internal/url/repository"
	"github.com/doanvuduy170921/quick-link/internal/url/usecase"
)

func TestShorten_Redirect_E2E(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	repo := repository.NewUrlRepository(testQueries)
	uc := usecase.NewUseCase(repo, testRedisCli)
	handler := handlerhttp.NewURLHandler(uc)

	router := gin.New()
	router.POST("/shorten", handler.ShortenURL)
	router.GET("/redirect/:code", handler.Redirect)
	shortenBody := `{"url":"https://real-integration-test.com"}`
	shortenReq := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(shortenBody))
	shortenReq.Header.Set("Content-Type", "application/json")
	shortenW := httptest.NewRecorder()

	router.ServeHTTP(shortenW, shortenReq)
	require.Equal(t, http.StatusOK, shortenW.Code)

	var shortenResp map[string]string
	require.NoError(t, json.Unmarshal(shortenW.Body.Bytes(), &shortenResp))
	code := shortenResp["code"]
	require.NotEmpty(t, code)

	redirectReq1 := httptest.NewRequest(http.MethodGet, "/redirect/"+code, nil)
	redirectW1 := httptest.NewRecorder()

	router.ServeHTTP(redirectW1, redirectReq1)
	require.Equal(t, http.StatusFound, redirectW1.Code)
	require.Equal(t, "https://real-integration-test.com", redirectW1.Header().Get("Location"))

	cachedURL, err := testRedisCli.Get(ctx, code)
	require.NoError(t, err)
	require.Equal(t, "https://real-integration-test.com", cachedURL)

	redirectReq2 := httptest.NewRequest(http.MethodGet, "/redirect/"+code, nil)
	redirectW2 := httptest.NewRecorder()

	router.ServeHTTP(redirectW2, redirectReq2)
	require.Equal(t, http.StatusFound, redirectW2.Code)
	require.Equal(t, "https://real-integration-test.com", redirectW2.Header().Get("Location"))

}
