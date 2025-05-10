package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
)

type Goods struct {
	ID         uint64    `json:"id"`
	LesseeID   uint64    `json:"lessee_id"`
	Name       string    `json:"name"`
	FinalPrice float64   `json:"final_price"`
	Price      float64   `json:"price"`
	Tags       []string  `json:"tags"`
	Avatar     string    `json:"avatar"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type GoodsSlice []Goods

func (a GoodsSlice) Len() int           { return len(a) }
func (a GoodsSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a GoodsSlice) Less(i, j int) bool { return a[i].UpdateTime.Before(a[j].UpdateTime) }

func (g *Goods) IsValid() (bool, string) {
	if g.ID == 0 {
		logrus.Errorln("goods id is 0")
		return false, ""
	}
	if g.LesseeID == 0 {
		return false, "非法租户"
	}
	if g.Name == "" {
		return false, "商品名为空"
	}
	if g.FinalPrice <= 0 {
		return false, "价格为空"
	}
	if g.Price <= 0 {
		return false, "原价为空"
	}
	if g.FinalPrice > g.Price {
		return false, "折后价需小于原价"
	}
	return true, ""
}

func (Goods) GetKey(lid, id uint64) string {
	if id == 0 {
		return fmt.Sprintf("goods/%d/", lid)
	}
	return fmt.Sprintf("goods/%d/%d", lid, id)
}
func (g *Goods) Save() error {
	return Set(g.GetKey(g.LesseeID, g.ID), g)
}

func (g Goods) GetGoods(lid uint64) (GoodsSlice, error) {
	prifix := g.GetKey(lid, 0)
	m, err := GetAllWithPrefix[Goods](prifix)
	if err != nil {
		return nil, err
	}
	var goods = make([]Goods, 0, len(m))
	for _, v := range m {
		goods = append(goods, v)
	}
	return goods, nil
}

func (g Goods) GetByID(lid, id uint64) (Goods, error) {
	key := g.GetKey(lid, id)
	var goods Goods
	err := Get(key, &goods)
	return goods, err
}

func (g *Goods) Update(lid, id uint64) error {
	return GetDB().Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(g.GetKey(lid, id)))
		if err != nil {
			return err
		}
		var old Goods
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &old)
		if err != nil {
			return err
		}
		if g.Name != "" {
			old.Name = g.Name
		}
		if g.FinalPrice != 0 {
			old.FinalPrice = g.FinalPrice
		}
		if g.Price != 0 {
			old.Price = g.Price
		}
		if len(g.Tags) > 0 {
			old.Tags = g.Tags
		}
		old.UpdateTime = time.Now()

		data, err = json.Marshal(old)
		if err != nil {
			return err
		}
		return txn.Set(item.KeyCopy(nil), data)
	})
}

func (g Goods) Delete(lid, id uint64) error {
	err := Delete(g.GetKey(lid, id))
	if err != nil {
		return err
	}
	DeleteImageByGoodsID(id)
	return nil
}
