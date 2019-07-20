# go-schedule-job
Simple cron like job scheduler 


```go
package main

import (
	"os"
	"time"
    "runtime"
    "github.com/etcd-io/bbolt"
    
    schedulejob "github.com/9glt/go-schedule-job"
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

var _ = (schedulejob.Store)(&DB{})

func TestMain(t *testing.T) {
	db, err := bbolt.Open(dbPath, 0755, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

    sc := schedulejob.New(&DB{db})
    
    // schedule job after 5seconds
    sc.Schedule(5, []byte("http://test1"))
    // true for auto delete items after send to channel
	ch, _ := sc.Worker(true)
	go func() {
	for1:
		for {
			select {
			case item, ok := <-ch:
				if ok {
                    // do you job with item.Payload
					continue
				}
				break for1
			}
		}
	}()
    
    runtime.Goexit()
}

```