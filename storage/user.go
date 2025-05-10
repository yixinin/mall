package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type UserKind string

const (
	Admin      UserKind = "admin"
	Manger     UserKind = "manager"
	Customer   UserKind = "customer"
	Technician UserKind = "tech"
)

type User struct {
	ID         uint64    `json:"id"`
	OpenID     string    `json:"open_id"`
	Kind       UserKind  `json:"kind"`
	Avatar     string    `json:"avatar"`
	Nickname   string    `json:"nickname"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (u *User) GetKey(id uint64) string {
	if id == 0 {
		return "user/id/"
	}
	return fmt.Sprintf("user/id/%d", id)
}
func (u *User) GetOpenKey(openid string) string {
	return fmt.Sprintf("user/open/%s", openid)
}

func (u *User) Save() error {
	key := u.GetKey(u.ID)
	openKey := u.GetOpenKey(u.OpenID)

	return GetDB().Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(u)
		if err != nil {
			return err
		}
		err = txn.Set([]byte(key), data)
		if err != nil {
			return err
		}
		data, _ = json.Marshal(u.ID)
		return txn.Set([]byte(openKey), data)
	})

}

func (u User) GetByID(id uint64) (User, error) {
	var user User
	err := Get(u.GetKey(id), &user)
	return user, err
}
func (u User) GetByOpenID(openid string) (User, error) {
	var id uint64
	err := Get(u.GetOpenKey(openid), &id)
	if err != nil {
		return User{}, err
	}
	return u.GetByID(id)
}

func (u User) GetUsers() ([]User, error) {
	prefix := u.GetKey(0)
	m, err := GetAllWithPrefix[User](prefix)
	if err != nil {
		return nil, err
	}
	var users = make([]User, 0, len(m))
	for _, v := range m {
		users = append(users, v)
	}
	return users, nil
}

func (u *User) Update(id uint64) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(u.GetKey(id)))
		if err != nil {
			return err
		}
		var old User
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &old)
		if err != nil {
			return err
		}
		if u.Nickname != "" {
			old.Nickname = u.Nickname
		}
		if u.Avatar != "" {
			old.Avatar = u.Avatar
		}
		if u.Kind != "" && u.Kind != old.Kind {
			old.Kind = u.Kind
		}
		old.UpdateTime = time.Now()

		data, err = json.Marshal(old)
		if err != nil {
			return err
		}
		return txn.Set(item.KeyCopy(nil), data)
	})
}

func (u User) Delete(id uint64) error {
	user, err := u.GetByID(id)
	if err != nil {
		return err
	}

	return GetDB().Update(func(txn *badger.Txn) error {
		if err := txn.Delete([]byte(u.GetKey(id))); err != nil {
			return err
		}
		return txn.Delete([]byte(u.GetOpenKey(user.OpenID)))
	})
}
