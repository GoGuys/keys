package ds

import (
	"context"
	"time"
)

// Change is used to add time stamped data to a collection.
// If this format changes, you should also change in firestore and other
// backends that don't directly use this struct on set.
type Change struct {
	Data      []byte    `json:"data" firestore:"data"`
	Timestamp time.Time `json:"ts" firestore:"ts"`
}

// Direction is ascending or descending.
type Direction string

const (
	// Ascending direction.
	Ascending Direction = "asc"
	// Descending direction.
	Descending Direction = "desc"
)

// Changes describes changes to a path.
type Changes interface {
	ChangeAdd(ctx context.Context, collection string, data []byte) (string, error)
	Changes(ctx context.Context, collection string, from time.Time, limit int, direction Direction) (ChangeIterator, error)
}
