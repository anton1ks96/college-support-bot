package domain

import "time"

type Message struct {
	ID        int64
	UserID    int64
	Text      string
	PhotoURLs []string
	CreatedAt time.Time
}
