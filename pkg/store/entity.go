package store

import "time"

type PBEntity struct {
	ID        string    `db:"id"`
	Payload   []byte    `db:"payload"`
	CreatedAt time.Time `db:"created_at"`
}
