package usecase

import (
	"context"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	"github.com/doanvuduy170921/quick-link/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestUseCase_Redirect(t *testing.T) {
	ctx := context.Background()
	futureTime := time.Now().Add(1 * time.Hour)
	//pastTime := time.Now().Add(-1 * time.Hour)
	tests := []struct {
		name      string
		code      string
		mockSetUp func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository)
		expectURL string
		expectErr bool
	}{
		{
			name: "success- found in redis(cache hit)",
			code: "123",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRe.EXPECT().Get(ctx, "123").Return("https://www.baidu.com", nil).Once()
			},
			expectURL: "https://www.baidu.com",
			expectErr: false},
		{
			name: "Success - Cache Miss, Found in DB",
			code: "abc123",
			mockSetUp: func(mockRe *mocks.MockRedisClient, mockRepo *mocks.MockUrlRepository) {
				mockRe.EXPECT().Get(ctx, "abc123").Return("", nil).Once()
				mockRepo.EXPECT().GetOriginalUrlByShortCode(ctx, "abc123").Return(db.GetOriginalUrlByShortCodeRow{
					OriginalUrl: "https://www.baidu.com",
					ExpiresAt:   &futureTime,
				}, nil).Once()
				mockRe.EXPECT().Set(ctx, "abc123", "https://www.baidu.com", mock.Anything).Return(nil).Once()
			},
			expectURL: "https://www.baidu.com",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRe := mocks.NewMockRedisClient(t)
			mockRepo := mocks.NewMockUrlRepository(t)

			tt.mockSetUp(mockRe, mockRepo)

			uc := NewUseCase(mockRepo, mockRe)

			url, err := uc.Redirect(ctx, tt.code)
			if (err != nil) != tt.expectErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.expectURL, url)
		})
	}
}
