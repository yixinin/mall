package handler

import (
	"mall/set"
	"mall/storage"
	"sort"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetOrders(c *gin.Context) {
	var req struct {
		ID       uint64              `form:"id"`
		Status   storage.OrderStatus `form:"status"`
		UserID   uint64              `form:"user_id"`
		LesseeID uint64              `form:"lessee_id"`
	}
	err := c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	lid := c.GetUint64("lid")
	if lid == 0 {
		RespMessage(c, "非法租户")
		return
	}
	uid := c.GetUint64("uid")
	logrus.Infof("lid:%d uid:%d get orders:%+v", lid, uid, req)

	user, err := storage.Model[storage.User]().GetByID(uid)
	if err != nil {
		logrus.Errorf("get user:%d error:%v", uid, err)
		RespInternalError(c, err)
		return
	}

	var orders storage.OrderSlice
	switch {
	case req.LesseeID != 0:
		lessee, err := storage.Model[storage.Lessee]().GetByID(lid)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		orders, err = storage.Model[storage.Order]().GetByLesseeID(req.LesseeID)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		switch user.Kind {
		case storage.Admin:
		case storage.Manger:
			if !set.From(lessee.Admins).Has(user.ID) {
				RespMessage(c, "no access")
				return
			}
		default:
			RespMessage(c, "no access")
			return
		}

	case req.ID != 0:
		order, err := storage.Model[storage.Order]().GetByID(lid, req.ID)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if order.User.ID != user.ID {
			RespInternalError(c, badger.ErrKeyNotFound)
			return
		}
		orders = append(orders, order)
	case req.UserID != 0:
		orders, err = storage.Model[storage.Order]().GetByUid(lid, req.UserID)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		switch user.Kind {
		case storage.Manger:
			lessee, err := storage.Model[storage.Lessee]().GetByID(lid)
			if err != nil {
				RespInternalError(c, err)
				return
			}
			if !set.From(lessee.Admins).Has(user.ID) {
				RespInternalError(c, badger.ErrKeyNotFound)
				return
			}
		case storage.Technician:
			RespInternalError(c, badger.ErrKeyNotFound)
			return
		case storage.Customer:
			RespInternalError(c, badger.ErrKeyNotFound)
			return
		case "":
			RespInternalError(c, badger.ErrKeyNotFound)
			return
		}
	case req.Status != "":
		if user.Kind != storage.Manger && user.Kind != storage.Admin {
			RespInternalError(c, badger.ErrKeyNotFound)
			return
		}
		orders, err = storage.Model[storage.Order]().GetByStatus(lid, req.Status)
		if err != nil {
			RespInternalError(c, err)
			return
		}
	default:
		orders, err = storage.Model[storage.Order]().GetByUid(lid, user.ID)
		if err != nil {
			RespInternalError(c, err)
			return
		}
	}
	sort.Sort(sort.Reverse(orders))
	Response(c, orders)
}

func (h *Handler) PostOrder(c *gin.Context) {
	var req struct {
		LesseeID uint64 `json:"lessee_id"`
		Goods    []struct {
			ID    uint64 `json:"id"`
			Count int    `json:"count"`
		} `json:"goods"`
		Time    string `json:"time"`
		Address string `json:"address"`
		Phone   string `json:"phone"`
	}
	err := c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	id, err := storage.GenID()
	if err != nil {
		RespInternalError(c, err)
		return
	}

	uid := c.GetUint64("uid")
	user, err := storage.Model[storage.User]().GetByID(uid)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	var now = time.Now()
	var order = &storage.Order{
		ID:       id,
		LesseeID: req.LesseeID,
		Status:   storage.Watting,
		Reverse: storage.OrderReverse{
			Time:    req.Time,
			Address: req.Address,
			Phone:   req.Phone,
		},
		User: storage.OrderUser{
			ID:       user.ID,
			Nickname: user.Nickname,
		},
		CreateTime: now,
		UpdateTime: now,
	}

	if order.LesseeID == 0 {
		order.LesseeID = c.GetUint64("lid")
	}

	if order.LesseeID == 0 {
		RespMessage(c, "非法租户")
		return
	}

	for _, g := range req.Goods {
		goods, err := storage.Model[storage.Goods]().GetByID(order.LesseeID, g.ID)
		if err != nil {
			RespBindError(c, err)
			return
		}
		order.Goods = append(order.Goods, storage.OrderGoods{
			ID:    g.ID,
			Count: g.Count,
			Price: goods.FinalPrice,
			Name:  goods.Name,
		})
		order.TotalPrice += goods.FinalPrice
	}

	valid, msg := order.IsValid()
	if !valid {
		RespMessage(c, msg)
		return
	}

	err = order.Save()
	if err != nil {
		RespBindError(c, err)
		return
	}
	Response(c, order)
}

func (h *Handler) PutOrder(c *gin.Context) {
	var req struct {
		ID       uint64 `uri:"id"`
		LesseeID uint64 `json:"lessee_id"`
		Time     string `json:"time"`
		Address  string `json:"address"`
		Phone    string `json:"phone"`
		Status   string `json:"status"`
	}
	err := c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	err = c.BindJSON(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}

	if req.LesseeID == 0 {
		req.LesseeID = c.GetUint64("lid")
	}
	if req.LesseeID == 0 {
		RespMessage(c, "非法租户")
		return
	}

	status := storage.OrderStatus(req.Status)
	if status != "" && !status.IsValid() {
		RespMessage(c, "status invalid")
		return
	}

	err = storage.Model[storage.Order]().Update(req.LesseeID, req.ID, req.Address, req.Time, req.Phone, status)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, req)
}
func (h *Handler) DeleteOrder(c *gin.Context) {
	var req struct {
		ID uint64 `uri:"id"`
	}
	err := c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	lid := c.GetUint64("lid")
	if lid == 0 {
		RespMessage(c, "非法租户")
		return
	}
	err = storage.Model[storage.Order]().Delete(lid, req.ID)
	if err != nil {
		RespInternalError(c, err)
		return
	}
}
