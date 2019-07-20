package schedulejob

import (
	"os"
	"testing"
	"time"

	"github.com/etcd-io/bbolt"
)

var (
	dbPath     = "./test-database.db-test"
	bucketJobs = []byte("jobs")
)

type DB struct {
	db *bbolt.DB
}

func (db *DB) First() ([]byte, []byte) {
	var key []byte
	var payload []byte

	db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketJobs)
		c := bucket.Cursor()
		k, v := c.First()
		key, payload = clone(k), clone(v)
		return nil
	})
	return key, payload
}
func (db *DB) Save(key, value []byte) {
	db.db.Update(func(tx *bbolt.Tx) error {
		bucket, _ := tx.CreateBucketIfNotExists(bucketJobs)
		bucket.Put(key, value)
		return nil
	})
}
func (db *DB) Delete(key []byte) {
	db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketJobs)
		bucket.Delete(key)
		return nil
	})
}

var _ = (Store)(&DB{})

func TestMain(t *testing.T) {
	db, err := bbolt.Open(dbPath, 0755, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	sc := New(&DB{db})
	sc.Schedule(5, []byte("http://test1"))
	sc.Schedule(7, []byte("http://test2"))
	ch, fn := sc.Worker(false)
	wait := make(chan struct{})
	go func() {
	for1:
		for {
			select {
			case item, ok := <-ch:
				if ok {
					// do your work with item.Payload
					item.Delete()
					// sc.Schedule(1, item.Payload)
					continue
				}
				break for1
			}
		}
		close(wait)
	}()
	go func() {
		time.Sleep(14 * time.Second)
		fn()
	}()
	<-wait
}

func clone(b []byte) []byte {
	r := make([]byte, len(b))
	copy(r, b)
	return r
}
