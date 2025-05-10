package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
)

type OrderStatus string

const (
	Watting  OrderStatus = "watting"
	Canceled OrderStatus = "canceled"
	Done     OrderStatus = "done"
)

func (o OrderStatus) Value() int {
	switch o {
	case Watting:
		return 3
	case Canceled:
		return 2
	case Done:
		return 1
	}
	return 0
}

// done < canceled < watting

func (o OrderStatus) Less(v OrderStatus) bool {

	return o.Value() < v.Value()
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case Watting:
	case Canceled:
	case Done:
	default:
		return false
	}
	return true
}

type OrderGoods struct {
	ID    uint64  `json:"id"`
	Price float64 `json:"price"`
	Name  string  `json:"name"`
	Count int     `json:"count"`
}

type OrderUser struct {
	ID       uint64 `json:"id"`
	Nickname string `json:"nickname"`
}

type OrderReverse struct {
	Time    string `json:"time"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type Order struct {
	ID         uint64       `json:"id"`
	LesseeID   uint64       `json:"lessee_id"`
	Status     OrderStatus  `json:"status"`
	Goods      []OrderGoods `json:"goods"`
	TotalPrice float64      `json:"total_price"`
	User       OrderUser    `json:"user"`
	Reverse    OrderReverse `json:"reverse"`
	CreateTime time.Time    `json:"create_time"`
	UpdateTime time.Time    `json:"update_time"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		CreateTime string `json:"create_time"`
		*Alias
	}{
		CreateTime: o.CreateTime.Format("2006-01-02 15:04:05"),
		Alias:      (*Alias)(&o),
	})
}

type OrderSlice []Order

func (a OrderSlice) Len() int      { return len(a) }
func (a OrderSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderSlice) Less(i, j int) bool {
	if a[i].Status == a[j].Status {
		return a[i].UpdateTime.Before(a[j].UpdateTime)
	}
	return a[i].Status.Less(a[j].Status)
}

func (o *Order) IsValid() (bool, string) {
	if o.ID == 0 {
		logrus.Errorln("goods id is 0")
		return false, ""
	}
	if o.LesseeID == 0 {
		return false, "非法租户"
	}
	if o.Status == "" {
		return false, "状态未设置"
	}
	if len(o.Goods) == 0 {
		return false, "商品为空"
	}
	if o.TotalPrice <= 0 {
		return false, "总价计算错误"
	}
	if o.User.ID == 0 {
		return false, "用户为空"
	}
	if o.User.Nickname == "" {
		return false, "用户名为空"
	}
	if o.Reverse.Address == "" {
		return false, "预约地址为空"
	}
	if o.Reverse.Phone == "" {
		return false, "联系电话为空"
	}
	if o.Reverse.Time == "" {
		return false, "预约时间为空"
	}
	return true, ""
}

func (Order) GetKey(lid, id uint64) string {
	if id == 0 {
		return fmt.Sprintf("order/%d/", lid)
	}
	return fmt.Sprintf("order/%d/%d", lid, id)
}
func (Order) GetUserOrdersKey(uid uint64) string {
	return fmt.Sprintf("user/orders/%d", uid)
}
func (o *Order) Save() error {
	userOrdersKey := o.GetUserOrdersKey(o.User.ID)
	orderKey := o.GetKey(o.LesseeID, o.ID)

	return GetDB().Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(userOrdersKey))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}

		var oids []uint64
		if err == nil {
			data, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			err = json.Unmarshal(data, &oids)
			if err != nil {
				return err
			}
		}
		oids = append(oids, o.ID)
		data, err := json.Marshal(oids)
		if err != nil {
			return err
		}
		err = txn.Set([]byte(userOrdersKey), data)
		if err != nil {
			return err
		}

		data, err = json.Marshal(o)
		if err != nil {
			return err
		}
		return txn.Set([]byte(orderKey), data)
	})
}

func (o Order) Update(lid, id uint64, address string, reverseTime string, phone string, status OrderStatus) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(o.GetKey(lid, id)))
		if err != nil {
			return err
		}
		var old Order
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &old)
		if err != nil {
			return err
		}

		if address != "" {
			old.Reverse.Address = address
		}
		if reverseTime != "" {
			old.Reverse.Time = reverseTime
		}
		if phone != "" {
			old.Reverse.Phone = phone
		}

		if status != "" && status != old.Status {
			old.Status = status
		}

		old.UpdateTime = time.Now()

		data, err = json.Marshal(old)
		if err != nil {
			return err
		}
		return txn.Set(item.KeyCopy(nil), data)
	})
}

func (o Order) Delete(lid, id uint64) error {
	return Delete(o.GetKey(lid, id))
}

func (o Order) GetByUid(lid, uid uint64) ([]Order, error) {
	userOrdersKey := o.GetUserOrdersKey(uid)
	var oids []uint64
	err := Get(userOrdersKey, &oids)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var orders []Order
	for _, id := range oids {
		var order Order
		key := o.GetKey(lid, id)
		err := Get(key, &order)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return nil, err
		}
		if order.ID != 0 {
			orders = append(orders, order)
		}
	}
	return orders, nil
}
func (o Order) GetByID(lid, id uint64) (Order, error) {
	var order Order
	key := o.GetKey(lid, id)
	err := Get(key, &order)
	return order, err
}

func (o Order) GetByLesseeID(lid uint64) ([]Order, error) {
	prefix := o.GetKey(lid, 0)
	m, err := GetAllWithPrefix[Order](prefix)
	if err != nil {
		return nil, err
	}
	var orders = make([]Order, 0, len(m))
	for _, v := range m {
		orders = append(orders, v)
	}
	return orders, nil
}
func (o Order) GetByStatus(lid uint64, status OrderStatus) ([]Order, error) {
	m, err := GetAllWithPrefix[Order](o.GetKey(lid, 0))
	if err != nil {
		return nil, err
	}
	var orders = make([]Order, 0, len(m))
	for _, v := range m {
		if v.Status == status {
			orders = append(orders, v)
		}
	}
	return orders, nil
}
