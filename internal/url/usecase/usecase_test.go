package usecase_test

import (
	"context"
	"errors"
	"github.com/doanvuduy170921/quick-link/internal/url/usecase"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	"github.com/doanvuduy170921/quick-link/internal/mocks"
	urlErr "github.com/doanvuduy170921/quick-link/internal/url/error"
)

// ==================== Redirect Tests ====================

func TestUseCase_Redirect(t *testing.T) {
	ctx := context.Background()
	futureTime := time.Now().Add(1 * time.Hour)
	pastTime := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name          string
		code          string
		mockSetUp     func(*mocks.MockRedisClient, *mocks.MockUrlRepository)
		expectURL     string
		expectErr     bool
		expectErrType urlErr.ErrorType
	}{
		{
			name: "cache hit - redis returns url",
			code: "abc123",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRe.EXPECT().
					Get(ctx, "abc123").
					Return("https://example.com", nil).
					Once()
			},
			expectURL: "https://example.com",
			expectErr: false,
		},
		{
			name: "cache miss - found in db, set cache",
			code: "xyz789",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRe.EXPECT().
					Get(ctx, "xyz789").
					Return("", nil).
					Once()
				mockRepo.EXPECT().
					GetOriginalUrlByShortCode(ctx, "xyz789").
					Return(db.GetOriginalUrlByShortCodeRow{
						OriginalUrl: "https://google.com",
						ExpiresAt:   &futureTime,
					}, nil).
					Once()
				mockRe.EXPECT().
					Set(ctx, "xyz789", "https://google.com", mock.Anything).
					Return(nil).
					Once()
			},
			expectURL: "https://google.com",
			expectErr: false,
		},
		{
			name: "code not found",
			code: "notexist",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRe.EXPECT().
					Get(ctx, "notexist").
					Return("", nil).
					Once()
				mockRepo.EXPECT().
					GetOriginalUrlByShortCode(ctx, "notexist").
					Return(db.GetOriginalUrlByShortCodeRow{}, pgx.ErrNoRows).
					Once()
			},
			expectErr:     true,
			expectErrType: urlErr.ErrNotFound,
		},
		{
			name: "link expired",
			code: "expired",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRe.EXPECT().
					Get(ctx, "expired").
					Return("", nil).
					Once()
				mockRepo.EXPECT().
					GetOriginalUrlByShortCode(ctx, "expired").
					Return(db.GetOriginalUrlByShortCodeRow{
						OriginalUrl: "https://old.com",
						ExpiresAt:   &pastTime,
					}, nil).
					Once()
			},
			expectErr:     true,
			expectErrType: urlErr.ErrNotFound,
		},
		{
			name: "code empty - validation error",
			code: "",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				// No mock calls expected
			},
			expectErr:     true,
			expectErrType: urlErr.ErrValidation,
		},
		{
			name: "redis error but db has data - still redirect",
			code: "fallback",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRe.EXPECT().
					Get(ctx, "fallback").
					Return("", errors.New("redis connection refused")).
					Once()
				mockRepo.EXPECT().
					GetOriginalUrlByShortCode(ctx, "fallback").
					Return(db.GetOriginalUrlByShortCodeRow{
						OriginalUrl: "https://fallback.com",
						ExpiresAt:   &futureTime,
					}, nil).
					Once()
				mockRe.EXPECT().
					Set(ctx, "fallback", "https://fallback.com", mock.Anything).
					Return(nil).
					Once()
			},
			expectURL: "https://fallback.com",
			expectErr: false,
		},
		{
			name: "redis set error - still return url",
			code: "nosync",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRe.EXPECT().
					Get(ctx, "nosync").
					Return("", nil).
					Once()
				mockRepo.EXPECT().
					GetOriginalUrlByShortCode(ctx, "nosync").
					Return(db.GetOriginalUrlByShortCodeRow{
						OriginalUrl: "https://nosync.com",
						ExpiresAt:   &futureTime,
					}, nil).
					Once()
				mockRe.EXPECT().
					Set(ctx, "nosync", "https://nosync.com", mock.Anything).
					Return(errors.New("redis set failed")).
					Once()
			},
			expectURL: "https://nosync.com",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRe := mocks.NewMockRedisClient(t)
			mockRepo := mocks.NewMockUrlRepository(t)

			tt.mockSetUp(mockRe, mockRepo)

			uc := usecase.NewUseCase(mockRepo, mockRe)
			url, err := uc.Redirect(ctx, tt.code)

			if tt.expectErr {
				require.Error(t, err)
				var urlError *urlErr.URLError
				require.True(t, errors.As(err, &urlError))
				require.Equal(t, tt.expectErrType, urlError.Type)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectURL, url)
			}
		})
	}
}

// ==================== GenerateKey Tests ====================

func TestUseCase_GenerateKey(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		url           string
		mockSetUp     func(*mocks.MockRedisClient, *mocks.MockUrlRepository)
		expectErr     bool
		expectErrType urlErr.ErrorType
	}{
		{
			name: "success - generate and cache",
			url:  "https://example.com",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRepo.EXPECT().
					CreateURL(ctx, mock.Anything).
					Return(db.Url{
						ShortCode:   "abc1234",
						OriginalUrl: "https://example.com",
					}, nil).
					Once()
				mockRe.EXPECT().
					Set(ctx, mock.Anything, "https://example.com", mock.Anything).
					Return(nil).
					Once()
			},
			expectErr: false,
		},
		{
			name: "url empty - validation error",
			url:  "",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				// No calls
			},
			expectErr:     true,
			expectErrType: urlErr.ErrValidation,
		},
		{
			name: "collision retry - success on 3rd try",
			url:  "https://retry.com",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				// 1st call: collision error
				mockRepo.EXPECT().
					CreateURL(ctx, mock.Anything).
					Return(db.Url{}, &pgconn.PgError{Code: "23505"}).
					Once()
				// 2nd call: collision error
				mockRepo.EXPECT().
					CreateURL(ctx, mock.Anything).
					Return(db.Url{}, &pgconn.PgError{Code: "23505"}).
					Once()
				// 3rd call: success
				mockRepo.EXPECT().
					CreateURL(ctx, mock.Anything).
					Return(db.Url{
						ShortCode:   "unique1",
						OriginalUrl: "https://retry.com",
					}, nil).
					Once()
				mockRe.EXPECT().
					Set(ctx, mock.Anything, "https://retry.com", mock.Anything).
					Return(nil).
					Once()
			},
			expectErr: false,
		},
		{
			name: "collision exhausted - after 3 retries",
			url:  "https://unlucky.com",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRepo.EXPECT().
					CreateURL(ctx, mock.Anything).
					Return(db.Url{}, &pgconn.PgError{Code: "23505"}).
					Times(3)
			},
			expectErr:     true,
			expectErrType: urlErr.ErrInternal,
		},
		{
			name: "db error (not unique violation) - no retry",
			url:  "https://error.com",
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
			name: "redis set error - still return code",
			url:  "https://nosync.com",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRepo.EXPECT().
					CreateURL(ctx, mock.Anything).
					Return(db.Url{
						ShortCode:   "fail1",
						OriginalUrl: "https://nosync.com",
					}, nil).
					Once()
				mockRe.EXPECT().
					Set(ctx, mock.Anything, "https://nosync.com", mock.Anything).
					Return(errors.New("redis down")).
					Once()
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRe := mocks.NewMockRedisClient(t)
			mockRepo := mocks.NewMockUrlRepository(t)

			tt.mockSetUp(mockRe, mockRepo)

			uc := usecase.NewUseCase(mockRepo, mockRe)
			code, err := uc.GenerateKey(ctx, tt.url)

			if tt.expectErr {
				require.Error(t, err)
				var urlError *urlErr.URLError
				require.True(t, errors.As(err, &urlError))
				require.Equal(t, tt.expectErrType, urlError.Type)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, code)
				require.Len(t, code, 7)
			}
		})
	}
}
