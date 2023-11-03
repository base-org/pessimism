package common

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

const (
	dlqFullMsg  = "the dead letter queue is full with %d elements"
	dlqEmptyMsg = "the dead letter queue is empty"
)

// DLQ ... Dead Letter Queue construct
// Used to store ETL events
// that failed to be processed
type DLQ[E any] struct {
	size int
	dlq  []*E
}

// NewTransitDLQ ... Initializer
func NewTransitDLQ(size int) *DLQ[core.TransitData] {
	return &DLQ[core.TransitData]{
		size: size,
		dlq:  make([]*core.TransitData, 0, size),
	}
}

// Add ... Adds an entry to the DLQ if it is not full
func (d *DLQ[E]) Add(entry *E) error {
	if len(d.dlq) >= d.size {
		return fmt.Errorf(dlqFullMsg, d.size)
	}

	d.dlq = append(d.dlq, entry)
	return nil
}

// Pop ... Removes the first element from the DLQ,
// typically for re-processing
func (d *DLQ[E]) Pop() (*E, error) {
	if len(d.dlq) == 0 {
		return nil, fmt.Errorf(dlqEmptyMsg)
	}

	entry := d.dlq[0]
	d.dlq = d.dlq[1:]
	return entry, nil
}

// PopAll ... Removes all elements from the DLQ,
// typically for re-processing
func (d *DLQ[E]) PopAll() []*E {
	entries := d.dlq
	d.dlq = make([]*E, 0, d.size)
	return entries
}

// Empty ... Checks if the DLQ is empty
func (d *DLQ[E]) Empty() bool {
	return len(d.dlq) == 0
}

// Size ... Returns the size of the DLQ
func (d *DLQ[E]) Size() int {
	return len(d.dlq)
}

// Full ... Checks if the DLQ is full
func (d *DLQ[E]) Full() bool {
	return len(d.dlq) >= d.size
}
