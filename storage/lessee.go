package storage

import (
	"encoding/json"
	"fmt"
	"mall/set"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
)

type LesseeStatus string

const (
	Disabled LesseeStatus = "disabled"
	Enabled  LesseeStatus = "enabled"
)

type Lessee struct {
	ID         uint64       `json:"id"`
	Admins     []uint64     `json:"admins"`
	Name       string       `json:"name"`
	Status     LesseeStatus `json:"enable"`
	CreateTime time.Time    `json:"create_time"`
	UpdateTime time.Time    `json:"update_time"`
}

func (l *Lessee) IsValid() (bool, string) {
	if l.ID == 0 {
		logrus.Errorln("lessee id is 0")
		return false, ""
	}
	if l.Name == "" {
		return false, "租户名为空"
	}
	if l.Status == "" {
		return false, "状态未设置"
	}
	return true, ""
}

func (Lessee) GetKey(id uint64) string {
	if id == 0 {
		return "lessee/"
	}
	return fmt.Sprintf("lessee/%d", id)
}
func (l *Lessee) Save() error {
	return Set(l.GetKey(l.ID), l)
}

func (l Lessee) GetByID(id uint64) (Lessee, error) {
	var lessee Lessee
	err := Get(l.GetKey(id), &lessee)
	return lessee, err
}

func (l Lessee) Update(id uint64, name string, status LesseeStatus, add []uint64, del []uint64) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(l.GetKey(id)))
		if err != nil {
			return err
		}
		var old Lessee
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &old)
		if err != nil {
			return err
		}

		if status != "" {
			old.Status = status
		}

		if name != "" {
			old.Name = name
		}
		admins := set.From(old.Admins).Add(add...).Del(del...).ToSlice()
		old.Admins = admins

		old.UpdateTime = time.Now()

		data, err = json.Marshal(old)
		if err != nil {
			return err
		}
		return txn.Set(item.KeyCopy(nil), data)
	})
}

func (l Lessee) GetLessees() ([]Lessee, error) {
	prefix := l.GetKey(0)
	m, err := GetAllWithPrefix[Lessee](prefix)
	if err != nil {
		return nil, err
	}
	var ls = make([]Lessee, 0, len(m))
	for _, v := range m {
		ls = append(ls, v)
	}
	return ls, nil
}

func (l Lessee) Delete(id uint64) error {
	return Delete(l.GetKey(id))
}
