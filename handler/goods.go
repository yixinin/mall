package handler

import (
	"mall/storage"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PostGoods(c *gin.Context) {
	var req struct {
		ID         uint64              `json:"id"`
		LesseeID   uint64              `json:"lessee_id"`
		Status     storage.GoodsStatus `json:"status"`
		Name       string              `json:"name"`
		Price      float64             `json:"price"`
		FinalPrice float64             `json:"final_price"`
		Tags       []string            `json:"tags"`
		Avatar     string              `json:"avatar"`
	}
	err := c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	if req.Avatar == "" {
		RespMessage(c, "商品图片为空")
		return
	}

	var now = time.Now()
	goods := storage.Goods{
		ID:         req.ID,
		LesseeID:   req.LesseeID,
		Name:       req.Name,
		Price:      req.Price,
		FinalPrice: req.FinalPrice,
		Status:     req.Status,
		Tags:       req.Tags,
		Avatar:     req.Avatar,
		CreateTime: now,
		UpdateTime: now,
	}
	if goods.LesseeID == 0 {
		goods.LesseeID = c.GetUint64("lid")
	}
	ok, msg := goods.IsValid()
	if !ok {
		RespMessage(c, msg)
		return
	}
	err = goods.Save()
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, goods)
}

func (h *Handler) GetGoodsList(c *gin.Context) {
	var req struct {
		Status storage.GoodsStatus `form:"status"`
	}
	err := c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	if req.Status != storage.Active {
		uid := c.GetUint64("uid")
		user, err := storage.Model[storage.User]().GetByID(uid)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if user.Kind != storage.Manger && user.Kind != storage.Admin {
			RespForbidden(c)
			return
		}
	}
	goods, err := storage.Model[storage.Goods]().GetGoods(c.GetUint64("lid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	var respGoods = make(storage.GoodsSlice, 0, len(goods))
	for i := range goods {
		if goods[i].Status == req.Status || req.Status == "" {
			respGoods = append(respGoods, goods[i])
		}
	}
	sort.Sort(sort.Reverse(respGoods))
	Response(c, respGoods)
}

type PreInfo struct {
	ID         uint64    `json:"id"`
	UpdateTime time.Time `json:"up"`
}

func (h *Handler) PreGetGoodsList(c *gin.Context) {
	var req struct {
		Status storage.GoodsStatus `form:"status"`
	}
	err := c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	if req.Status != storage.Active {
		uid := c.GetUint64("uid")
		user, err := storage.Model[storage.User]().GetByID(uid)
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if user.Kind != storage.Manger && user.Kind != storage.Admin {
			RespForbidden(c)
			return
		}
	}
	goods, err := storage.Model[storage.Goods]().GetGoods(c.GetUint64("lid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	var respGoods = make(storage.GoodsSlice, 0, len(goods))
	for i := range goods {
		if goods[i].Status == req.Status || req.Status == "" {
			respGoods = append(respGoods, goods[i])
		}
	}
	sort.Sort(sort.Reverse(respGoods))
	var infos = make([]PreInfo, 0, len(respGoods))
	for _, v := range respGoods {
		infos = append(infos, PreInfo{
			ID:         v.ID,
			UpdateTime: v.UpdateTime,
		})
	}
	Response(c, infos)
}
func (h *Handler) GetGoods(c *gin.Context) {
	var req struct {
		ID uint64 `uri:"id"`
	}
	err := c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	if req.ID == 0 {
		RespBindError(c, err)
		return
	}
	goods, err := storage.Model[storage.Goods]().GetByID(c.GetUint64("lid"), req.ID)
	if err != nil {
		RespInternalError(c, err)
		return
	}

	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusNoContent)
		c.Writer.Header().Set("x-up", storage.MarshalTime(goods.UpdateTime))
		return
	}
	Response(c, goods)
}

func (h *Handler) PutGoods(c *gin.Context) {
	var req struct {
		ID         uint64              `uri:"id"`
		Status     storage.GoodsStatus `json:"status"`
		Name       string              `json:"name"`
		Price      float64             `json:"price"`
		FinalPrice float64             `json:"final_price"`
		Tags       []string            `json:"tags"`
	}
	err := c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	err = c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	if req.ID == 0 {
		RespBindError(c, err)
		return
	}

	var now = time.Now()
	goods := storage.Goods{
		ID:         req.ID,
		Name:       req.Name,
		Status:     req.Status,
		Price:      req.Price,
		FinalPrice: req.FinalPrice,
		Tags:       req.Tags,
		UpdateTime: now,
	}
	if goods.LesseeID == 0 {
		goods.LesseeID = c.GetUint64("lid")
	}
	if goods.LesseeID == 0 {
		RespMessage(c, "非法租户")
		return
	}
	err = goods.Update(goods.LesseeID, req.ID)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, goods)
}
func (h *Handler) DeleteGoods(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		RespBindError(c, err)
		return
	}
	lid := c.GetUint64("lid")
	if lid == 0 {
		RespMessage(c, "非法租户")
		return
	}
	err = storage.Model[storage.Goods]().Delete(lid, id)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	key := storage.GetGoodsAvatarImageKey(id)
	storage.Delete(key)
}

func GetTags(s string) []string {
	tags := make([]string, 0)
	for _, v := range strings.Split(s, ",") {
		if v := strings.TrimSpace(v); len(v) > 0 {
			tags = append(tags, v)
		}
	}
	return tags
}
