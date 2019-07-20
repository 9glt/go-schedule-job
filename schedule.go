package schedulejob

import (
	"bytes"
	"encoding/binary"
	"time"
)

// Store interface for data  persists and read data from database
type Store interface {
	First() ([]byte, []byte)
	Save(key, value []byte)
	Delete(key []byte)
}

// New creates new schedule instance
func New(store Store) *Scheduler {
	s := &Scheduler{
		store: store,
	}
	return s
}

// Item struct
type Item struct {
	key     []byte
	Payload []byte
	s       Store
	deleted bool
}

// Delete from schedule key after done
func (i *Item) Delete() {
	if i.deleted {
		return
	}
	i.s.Delete(i.key)
}

// Scheduler main truct
type Scheduler struct {
	store Store
}

// Schedule takes seconds to delay job and payload
func (s *Scheduler) Schedule(after int64, payload []byte) {
	next := int64(time.Now().UnixNano() + (after * time.Second.Nanoseconds()))
	s.store.Save(itob(next), payload)
}

// Worker periodicaly reads store and found jobs pushes to channel. returns chan Item + cancel func
func (s *Scheduler) Worker(autodelete bool) (chan Item, func()) {
	ch := make(chan Item)
	worker := make(chan struct{})
	go func() {
	formain:
		for {
			time.Sleep(1 * time.Second)
			for {
				select {
				case _, ok := <-worker:
					if !ok {
						return
					}
				default:
				}

				key, payload := s.store.First()
				if bytes.Compare(key, []byte{}) == 0 {
					continue formain
				}
				if bytes.Compare(key, itob(time.Now().UnixNano())) > 0 {
					continue formain
				}
				item := Item{key, payload, s.store, false}

				ch <- item
				if autodelete {
					item.Delete()
				}
			}
		}
	}()
	return ch, func() { close(worker); time.Sleep(1 * time.Second); close(ch) }
}

func itob(in int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(in))
	return b
}
