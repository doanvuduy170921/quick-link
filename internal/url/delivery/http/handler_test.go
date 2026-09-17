package http

import (
	"context"
	"errors"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	"github.com/doanvuduy170921/quick-link/internal/mocks"
	urlErr "github.com/doanvuduy170921/quick-link/internal/url/error"
	"github.com/doanvuduy170921/quick-link/internal/url/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func pgUniqueViolationErr() error {
	return &pgconn.PgError{Code: "23505"}
}

func TestUseCase_GenerateKey_CustomAlias(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		url           string
		customAlias   string
		userID        *int64
		mockSetUp     func(*mocks.MockRedisClient, *mocks.MockUrlRepository)
		expectCode    string
		expectErr     bool
		expectErrType urlErr.ErrorType
	}{
		{
			name:        "custom alias - success with logged in user",
			url:         "https://example.com",
			customAlias: "mylink",
			userID:      int64Ptr(42),
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRepo.EXPECT().
					CreateURL(ctx, mock.MatchedBy(func(p db.CreateURLParams) bool {
						return p.ShortCode == "mylink" &&
							p.UserID.Valid == true &&
							p.UserID.Int64 == 42
					})).
					Return(db.Url{
						ShortCode:   "mylink",
						OriginalUrl: "https://example.com",
					}, nil).
					Once()
				mockRe.EXPECT().
					Set(ctx, "mylink", "https://example.com", mock.Anything).
					Return(nil).
					Once()
			},
			expectCode: "mylink",
			expectErr:  false,
		},
		{
			name:        "custom alias - success with anonymous user (userID nil)",
			url:         "https://example.com",
			customAlias: "publiclink",
			userID:      nil,
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRepo.EXPECT().
					CreateURL(ctx, mock.MatchedBy(func(p db.CreateURLParams) bool {
						// Verify: userID nil PHẢI convert thành pgtype.Int8{Valid: false}
						return p.ShortCode == "publiclink" && p.UserID.Valid == false
					})).
					Return(db.Url{
						ShortCode:   "publiclink",
						OriginalUrl: "https://example.com",
					}, nil).
					Once()
				mockRe.EXPECT().
					Set(ctx, "publiclink", "https://example.com", mock.Anything).
					Return(nil).
					Once()
			},
			expectCode: "publiclink",
			expectErr:  false,
		},
		{
			name:        "custom alias - already taken (unique violation, NO retry)",
			url:         "https://example.com",
			customAlias: "taken",
			userID:      int64Ptr(1),
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRepo.EXPECT().
					CreateURL(ctx, mock.Anything).
					Return(db.Url{}, pgUniqueViolationErr()).
					Once() // ← Once() tự động verify KHÔNG có lần gọi thứ 2 (không retry)
			},
			expectErr:     true,
			expectErrType: urlErr.ErrValidation,
		},
		{
			name:        "custom alias - other db error (not unique violation)",
			url:         "https://example.com",
			customAlias: "somealias",
			userID:      int64Ptr(1),
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRepo.EXPECT().
					CreateURL(ctx, mock.Anything).
					Return(db.Url{}, errors.New("connection timeout")).
					Once()
			},
			expectErr:     true,
			expectErrType: urlErr.ErrInternal,
		},
		{
			name:        "custom alias - redis set error, still return code",
			url:         "https://example.com",
			customAlias: "resilient",
			userID:      int64Ptr(1),
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRepo.EXPECT().
					CreateURL(ctx, mock.Anything).
					Return(db.Url{
						ShortCode:   "resilient",
						OriginalUrl: "https://example.com",
					}, nil).
					Once()
				mockRe.EXPECT().
					Set(ctx, "resilient", "https://example.com", mock.Anything).
					Return(errors.New("redis down")).
					Once()
			},
			expectCode: "resilient",
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRe := mocks.NewMockRedisClient(t)
			mockRepo := mocks.NewMockUrlRepository(t)

			tt.mockSetUp(mockRe, mockRepo)

			uc := usecase.NewUseCase(mockRepo, mockRe)
			code, err := uc.GenerateKey(ctx, tt.url, tt.customAlias, tt.userID)

			if tt.expectErr {
				require.Error(t, err)
				var urlError *urlErr.URLError
				require.True(t, errors.As(err, &urlError))
				require.Equal(t, tt.expectErrType, urlError.Type)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectCode, code)
			}
		})
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func TestURLHandler_Redirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		code              string
		mockSetUp         func(*mocks.MockUseCase)
		ExpectStatusCode  int
		expectTrackCalled bool
	}{
		{
			name: "success",
			code: "abcd123",
			mockSetUp: func(m *mocks.MockUseCase) {
				m.EXPECT().Redirect(mock.Anything, "abcd123").Return("http://google.com", nil).Once()
			},
			expectTrackCalled: true,
			ExpectStatusCode:  http.StatusFound,
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
			mockTracker := mocks.NewMockClickTracker(t)
			if tc.expectTrackCalled {
				mockTracker.EXPECT().Track(mock.Anything).Once()
			}
			s := NewURLHandler(mockUC, mockTracker)
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
