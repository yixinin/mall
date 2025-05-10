package storage

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
)

var db *badger.DB
var seq *badger.Sequence

func Init(dbpath string) error {
	opt := badger.DefaultOptions(dbpath)
	opt.WithInMemory(dbpath == "")
	opt = opt.WithLogger(logrus.StandardLogger())
	opt.ValueLogFileSize = int64(300) << 20 // max valueLog 300M
	var err error

	db, err = badger.Open(opt)
	if err != nil {
		return err
	}
	seq, err = db.GetSequence([]byte("id.gen"), 1)
	if err != nil {
		return err
	}

	go func() {
		defer seq.Release()
		tk := time.NewTicker(10 * time.Minute)
		for {
			select {
			// case <-cmd.Context().Done():
			// 	return
			case <-tk.C:
				db.RunValueLogGC(0.5)
			}
		}
	}()
	return err
}

func GetDB() *badger.DB {
	return db
}

func Close() {
	if seq != nil {
		seq.Release()
	}
	if db != nil {
		db.Close()
	}
}

func Model[T any]() T {
	var i T
	return i
}

func GenID() (uint64, error) {
	id, err := seq.Next()
	if err != nil {
		return 0, err
	}
	return id + 1024, nil
}

func Set(key string, data any, ttl ...time.Duration) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if len(ttl) > 0 && ttl[0] > 0 {
			e := badger.NewEntry([]byte(key), data).WithTTL(ttl[0])
			return txn.SetEntry(e)
		}
		err = txn.Set([]byte(key), data)
		if err != nil {
			return err
		}
		return nil
	})
}

func Get(key string, value any) error {
	return GetDB().View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, value)
	})
}

func Delete(key string) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		return nil
	})
}

func GetAllWithPrefix[T any](match string) (map[string]T, error) {
	var vals = make(map[string]T, 0)
	err := GetDB().View(func(txn *badger.Txn) error {
		var opts = badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(match)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if len(val) > 0 {
				var data T
				err := json.Unmarshal(val, &data)
				if err != nil {
					return err
				}
				vals[string(key)] = data
			}
		}
		return nil
	})
	return vals, err
}

func DeleteAllWithPrefix(match string) error {
	var keys = make([]string, 0)
	err := GetDB().View(func(txn *badger.Txn) error {
		var opts = badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(match)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			keys = append(keys, string(key))
		}
		return nil
	})
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return GetDB().Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			txn.Delete([]byte(key))
		}
		return nil
	})
}

func Scan[T any](prefix string, startCursor, endCursor string, limit int) (vals map[string]T, cursor string, err error) {
	vals = make(map[string]T)
	err = GetDB().View(func(txn *badger.Txn) error {
		var opts = badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(prefix)
		var total int
		for it.Seek([]byte(startCursor)); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.KeyCopy(nil))
			cursor = key
			if endCursor != "" && key > endCursor {
				break
			}
			if limit > 0 && total > limit {
				break
			}
			total++
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if len(val) > 0 {
				var data T
				err := json.Unmarshal(val, &data)
				if err != nil {
					return err
				}
				vals[key] = data
			}
		}
		return nil
	})
	return
}

func Range[T any](prefix string, startCursor, endCursor string, limit int, f func(key string, val T) bool) error {
	return GetDB().View(func(txn *badger.Txn) error {
		var opts = badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(prefix)
		var total int
		for it.Seek([]byte(startCursor)); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.KeyCopy(nil))
			if endCursor != "" && key > endCursor {
				break
			}
			if limit > 0 && total > limit {
				break
			}
			total++
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			if len(val) > 0 {
				var data T
				err := json.Unmarshal(val, &data)
				if err != nil {
					return err
				}
				if !f(key, data) {
					return nil
				}
			}
		}
		return nil
	})
}
