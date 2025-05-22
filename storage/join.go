package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type Join struct {
	ID         uint64     `json:"id"`
	User       SimpleUser `json:"user"`
	Lessee     IdName     `json:"lessee"`
	Status     JoinStatus `json:"status"`
	CreateTime time.Time  `json:"create_time"`
	UpdateTime time.Time  `json:"update_time"`
}
type JoinStatus string

const (
	WattingJoin JoinStatus = "watting"
	AcceptJoin  JoinStatus = "accept"
	RefuseJoin  JoinStatus = "refuse"
)

type IdName struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func (j *Join) GetKey(lid, id uint64) string {
	if id == 0 {
		return fmt.Sprintf("join/id/%d/", lid)
	}
	return fmt.Sprintf("join/id/%d/%d", lid, id)
}

func (j *Join) Save() error {
	key := j.GetKey(j.Lessee.Id, j.ID)

	return GetDB().Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(j)
		if err != nil {
			return err
		}
		return txn.Set([]byte(key), data)
	})

}

func (j Join) GetByID(lid, id uint64) (Join, error) {
	var join Join
	err := Get(j.GetKey(lid, id), &join)
	return join, err
}

func (j Join) GetJoins(lessee uint64) ([]Join, error) {
	prefix := j.GetKey(lessee, 0)
	m, err := GetAllWithPrefix[Join](prefix)
	if err != nil {
		return nil, err
	}
	var users = make([]Join, 0, len(m))
	for _, v := range m {
		users = append(users, v)
	}
	return users, nil
}

func (j Join) Update(lid, id uint64, status JoinStatus) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		key := j.GetKey(lid, id)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		var old Join
		err = json.Unmarshal(data, &old)
		if err != nil {
			return err
		}
		old.Status = status
		old.UpdateTime = time.Now()
		data, err = json.Marshal(old)
		if err != nil {
			return err
		}
		return txn.Set([]byte(key), data)
	})
}

func (j Join) Delete(lid, id uint64) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(j.GetKey(lid, id)))
	})
}
