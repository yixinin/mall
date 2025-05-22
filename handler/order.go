package handler

import (
	"context"
	"fmt"
	"mall/set"
	"mall/storage"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/basicService/subscribeMessage/request"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/power"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetOrders(c *gin.Context) {
	var req struct {
		Status   storage.OrderStatus `form:"status"`
		LesseeID uint64              `form:"lessee_id"`
		Manage   bool                `form:"manage"`
	}
	err := c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	lid := c.GetUint64("lid")
	if req.LesseeID == 0 {
		req.LesseeID = lid
	}
	if req.LesseeID == 0 {
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

	if req.Manage {
		oo, err := storage.Model[storage.Order]().GetByLesseeID(req.LesseeID, req.Status)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		orders = make(storage.OrderSlice, 0, len(oo))
		lessee, err := storage.Model[storage.Lessee]().GetByID(req.LesseeID)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if set.From(lessee.Admins).Has(uid) || user.Kind == storage.Admin {
			orders = oo
		} else if set.From(lessee.Techs).Has(uid) {
			for i := range oo {
				if oo[i].Status == storage.Watting || oo[i].Tech.ID == uid {
					orders = append(orders, oo[i])
				}
			}
		}

	} else {
		orders, err = storage.Model[storage.Order]().GetByUid(lid, user.ID, req.Status)
		if err != nil {
			RespInternalError(c, err)
			return
		}
	}

	sort.Sort(sort.Reverse(orders))
	if len(orders) > 100 {
		orders = orders[:100]
	}
	Response(c, orders)
}

func (h *Handler) PreGetOrders(c *gin.Context) {
	var req struct {
		Status   storage.OrderStatus `form:"status"`
		LesseeID uint64              `form:"lessee_id"`
		Manage   bool                `form:"manage"`
	}
	err := c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	lid := c.GetUint64("lid")
	if req.LesseeID == 0 {
		req.LesseeID = lid
	}
	if req.LesseeID == 0 {
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

	if req.Manage {
		oo, err := storage.Model[storage.Order]().GetByLesseeID(req.LesseeID, req.Status)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		orders = make(storage.OrderSlice, 0, len(oo))
		lessee, err := storage.Model[storage.Lessee]().GetByID(req.LesseeID)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if set.From(lessee.Admins).Has(uid) || user.Kind == storage.Admin {
			orders = oo
		} else if set.From(lessee.Techs).Has(uid) {
			for i := range oo {
				if oo[i].Status == storage.Watting || oo[i].Tech.ID == uid {
					orders = append(orders, oo[i])
				}
			}
		}

	} else {
		orders, err = storage.Model[storage.Order]().GetByUid(lid, user.ID, req.Status)
		if err != nil {
			RespInternalError(c, err)
			return
		}
	}

	sort.Sort(sort.Reverse(orders))
	var infos = make([]PreInfo, 0, len(orders))
	for _, v := range orders {
		infos = append(infos, PreInfo{
			ID:         v.ID,
			UpdateTime: v.UpdateTime,
		})
	}
	Response(c, infos)
}

func (h *Handler) GetOrder(c *gin.Context) {
	var req struct {
		ID uint64 `uri:"id"`
	}
	err := c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	uid := c.GetUint64("uid")
	lid := c.GetUint64("lid")

	order, err := storage.Model[storage.Order]().GetByID(lid, req.ID)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusNoContent)
		c.Writer.Header().Set("x-up", storage.MarshalTime(order.UpdateTime))
		return
	}
	if order.User.ID == uid {
		Response(c, order)
		return
	}

	user, err := storage.Model[storage.User]().GetByID(uid)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	switch user.Kind {
	case storage.Admin:
		Response(c, order)
		return
	case storage.Manger:
		lessee, err := storage.Model[storage.Lessee]().GetByID(lid)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if set.From(lessee.Admins).Has(uid) {
			Response(c, order)
			return
		}
	case storage.Technician:
		lessee, err := storage.Model[storage.Lessee]().GetByID(lid)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if set.From(lessee.Techs).Has(uid) {
			Response(c, order)
			return
		}
	}

	RespMessage(c, "not allow")
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
		Remark  string `json:"remark"`
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
	if req.LesseeID == 0 {
		req.LesseeID = c.GetUint64("lid")
	}

	uid := c.GetUint64("uid")
	user, err := storage.Model[storage.User]().GetByID(uid)
	if err != nil {
		RespInternalError(c, err)
		return
	}

	lessee, err := storage.Model[storage.Lessee]().GetByID(req.LesseeID)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	var notifyUser string
	if len(lessee.Admins) > 0 {
		manager, err := storage.Model[storage.User]().GetByID(lessee.Admins[0])
		if err != nil {
			RespInternalError(c, err)
			return
		}
		notifyUser = manager.OpenID
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
			Remark:  req.Remark,
		},
		User: storage.SimpleUser{
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
			RespInternalError(c, err)
			return
		}
		order.Goods = append(order.Goods, storage.OrderGoods{
			ID:    g.ID,
			Count: g.Count,
			Price: goods.FinalPrice,
			Name:  goods.Name,
		})
		order.TotalPrice += goods.FinalPrice * float64(g.Count)
	}

	valid, msg := order.IsValid()
	if !valid {
		RespMessage(c, msg)
		return
	}

	err = order.Save()
	if err != nil {
		RespInternalError(c, err)
		return
	}

	// 通知师傅
	if notifyUser != "" {
		// go h.SendNewOrderMessage(notifyUser, order)
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

	user, err := storage.Model[storage.User]().GetByID(c.GetUint64("uid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	order, err := storage.Model[storage.Order]().GetByID(req.LesseeID, req.ID)
	if err != nil {
		RespInternalError(c, err)
		return
	}

	lessee, err := storage.Model[storage.Lessee]().GetByID(req.LesseeID)
	if err != nil {
		RespInternalError(c, err)
		return
	}

	switch status {
	case storage.Comfirm, storage.Done:
		switch user.Kind {
		case storage.Customer:
			RespForbidden(c)
			return
		case storage.Technician, storage.Manger:
			if !set.From(lessee.Admins).Has(user.ID) && !set.From(lessee.Techs).Has(user.ID) {
				RespForbidden(c)
				return
			}
		}
	case storage.Canceled:
		switch user.Kind {
		case storage.Customer:
			if user.ID != order.User.ID {
				RespForbidden(c)
				return
			}
		case storage.Technician, storage.Manger:
			if !set.From(lessee.Admins).Has(user.ID) && !set.From(lessee.Techs).Has(user.ID) {
				RespForbidden(c)
				return
			}
		}
	}

	old, err := storage.Model[storage.Order]().Update(req.LesseeID, req.ID, req.Address, req.Time, req.Phone, status)
	if err != nil {
		RespInternalError(c, err)
		return
	}

	if old.Status != status {
		switch status {
		case storage.Comfirm:
			// 通知客户
		case storage.Canceled:
			// 通知师傅
		}
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

func (h *Handler) SendNewOrderMessage(openid string, order *storage.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var infos []string
	for _, v := range order.Goods {
		infos = append(infos, fmt.Sprintf("%s x %d", v.Name, v.Count))
	}

	data := &power.HashMap{
		"thing1": map[string]string{
			"value": order.Reverse.Address,
		},
		"thing4": map[string]string{
			"value": strings.Join(infos, ", "),
		},
		"thing7": map[string]string{
			"value": order.Reverse.Time,
		},
		"phone_number13": map[string]string{
			"value": order.Reverse.Phone,
		},
		"thing9": map[string]string{
			"value": order.Reverse.Remark,
		},
	}
	result, err := h.wxApp.SubscribeMessage.Send(ctx, &request.RequestSubscribeMessageSend{
		ToUser:           openid,
		TemplateID:       "Z8rbyG0DzQu8d4KYqgABsw8YpbfQ5yHFcRrh86eECS4",
		Page:             "pages/order/order",
		Data:             data,
		Lang:             "zh_CN",
		MiniProgramState: "formal",
	})
	if err != nil {
		return err
	}
	logrus.Infof("send new order notify code: %d, msg:", result.Code, result.ResultMsg)
	return nil
}

func (h *Handler) SendConfirmOrderMessage(openid, phone string, order *storage.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var infos []string
	for _, v := range order.Goods {
		infos = append(infos, fmt.Sprintf("%s x %d", v.Name, v.Count))
	}

	data := &power.HashMap{
		"name3": map[string]string{
			"value": order.Reverse.Address,
		},
		"thing9": map[string]string{
			"value": strings.Join(infos, ", "),
		},
		"date8": map[string]string{
			"value": order.Reverse.Time,
		},
		"phone_number6": map[string]string{
			"value": phone,
		},
	}
	result, err := h.wxApp.SubscribeMessage.Send(ctx, &request.RequestSubscribeMessageSend{
		ToUser:           openid,
		TemplateID:       "AQEXvDA2KcUTDuPX7jBtO1BBOwf0_0wyKh-QuAK-sY0",
		Page:             "pages/order/order",
		Data:             data,
		Lang:             "zh_CN",
		MiniProgramState: "formal",
	})
	if err != nil {
		return err
	}
	logrus.Infof("send new order notify code: %s, msg:%s", result.Code, result.ResultMsg)
	return nil
}

func (h *Handler) SendCancelOrderMessage(openid string, order *storage.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	data := &power.HashMap{
		"thing1": map[string]string{
			"value": order.Reverse.Address,
		},
		"date10": map[string]string{
			"value": order.Reverse.Time,
		},
	}
	result, err := h.wxApp.SubscribeMessage.Send(ctx, &request.RequestSubscribeMessageSend{
		ToUser:           openid,
		TemplateID:       "WR3oyAQ_sgIXOBd3gBsMWWi1c-gHJ03rAc-zJc9978s",
		Page:             "pages/order/order",
		Data:             data,
		Lang:             "zh_CN",
		MiniProgramState: "formal",
	})
	if err != nil {
		return err
	}
	logrus.Infof("send new order notify code: %s, msg:%s", result.Code, result.ResultMsg)
	return nil
}

func (h *Handler) SendUpdateTimeOrderMessage(openid string, order *storage.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	data := &power.HashMap{
		"thing1": map[string]string{
			"value": order.Reverse.Address,
		},
		"date10": map[string]string{
			"value": order.Reverse.Time,
		},
	}
	result, err := h.wxApp.SubscribeMessage.Send(ctx, &request.RequestSubscribeMessageSend{
		ToUser:           openid,
		TemplateID:       "WR3oyAQ_sgIXOBd3gBsMWWi1c-gHJ03rAc-zJc9978s",
		Page:             "pages/order/order",
		Data:             data,
		Lang:             "zh_CN",
		MiniProgramState: "formal",
	})
	if err != nil {
		return err
	}
	logrus.Infof("send new order notify code: %s, msg:%s", result.Code, result.ResultMsg)
	return nil
}
