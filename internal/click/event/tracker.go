package event

import (
	"context"
	"github.com/doanvuduy170921/quick-link/internal/click/repository"
	db "github.com/doanvuduy170921/quick-link/internal/infrastructure/db/sqlc"
	"log"
	"time"
)

type ClickEvent struct {
	ShortCode string
	ClickAt   time.Time
}

type ClickTracker struct {
	Event     chan ClickEvent
	ClickRepo repository.ClickRepository
}

func NewClickTracker(ClickRepo repository.ClickRepository, bufferSize int) *ClickTracker {
	return &ClickTracker{
		Event:     make(chan ClickEvent, bufferSize),
		ClickRepo: ClickRepo,
	}
}

func (c *ClickTracker) Track(code string) {
	select {
	case c.Event <- ClickEvent{ShortCode: code, ClickAt: time.Now()}:

	default:
		log.Printf("click tracker buffer full, dropping event for code=%s", code)
	}
}
func (c *ClickTracker) StartWorker(ctx context.Context, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go c.worker(ctx)
	}
}

func (c *ClickTracker) worker(ctx context.Context) {
	batchSize := 10
	batch := []db.BatchInsertClickParams{}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case event := <-c.Event:
			batch = append(batch, db.BatchInsertClickParams{ShortCode: event.ShortCode, ClickedAt: event.ClickAt})
			if len(batch) >= batchSize {
				c.flush(&batch)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				c.flush(&batch)
			}
		case <-ctx.Done():
			c.flush(&batch)
			return
		}
	}

}

func (c *ClickTracker) flush(batch *[]db.BatchInsertClickParams) {
	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := c.ClickRepo.BatchInsertClick(flushCtx, *batch)
	if err != nil {
		log.Printf("WARNING: lost %d click events due to insert error: %v", len(*batch), err)
	} else {
		log.Printf("flushed %d click events", rows)
	}
	*batch = (*batch)[:0]
}
