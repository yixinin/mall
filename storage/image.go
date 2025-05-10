package storage

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

func GetGoodsAvatarImageKey(id uint64) string {
	return fmt.Sprintf("img/avatar/goods/%d", id)
}

func SaveImage(key string, data []byte) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func GetImage(key string) ([]byte, error) {
	var data []byte
	err := GetDB().View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		data = val
		return err
	})
	return data, err
}

func DeleteImageByGoodsID(id uint64) error {
	return Delete(GetGoodsAvatarImageKey(id))
}
