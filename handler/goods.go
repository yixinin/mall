package handler

import (
	"fmt"
	"io"
	"mall/storage"
	"mime/multipart"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PostGoods(c *gin.Context) {
	var req struct {
		LesseeID   uint64                `form:"lessee_id"`
		Name       string                `form:"name"`
		Price      float64               `form:"price"`
		FinalPrice float64               `form:"final_price"`
		Tags       string                `form:"tags"`
		Avatar     *multipart.FileHeader `form:"avatar"`
	}
	err := c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	if req.Avatar == nil {
		RespMessage(c, "商品图片为空")
		return
	}
	fs, err := req.Avatar.Open()
	if err != nil {
		RespBindError(c, err)
		return
	}
	defer fs.Close()

	img, err := io.ReadAll(fs)
	if err != nil {
		RespBindError(c, err)
		return
	}
	id, err := storage.GenID()
	if err != nil {
		RespInternalError(c, err)
		return
	}

	imageKey := storage.GetGoodsAvatarImageKey(id)
	err = storage.SaveImage(imageKey, img)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	var now = time.Now()
	goods := storage.Goods{
		ID:         id,
		LesseeID:   req.LesseeID,
		Name:       req.Name,
		Price:      req.Price,
		FinalPrice: req.FinalPrice,
		Tags:       GetTags(req.Tags),
		Avatar:     fmt.Sprintf("/%s", imageKey),
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
	goods, err := storage.Model[storage.Goods]().GetGoods(c.GetUint64("lid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	sort.Sort(sort.Reverse(goods))
	Response(c, goods)
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
	Response(c, goods)
}

func (h *Handler) PutGoods(c *gin.Context) {
	var req struct {
		ID         uint64                `uri:"id"`
		LesseeID   uint64                `form:"lessee_id"`
		Name       string                `form:"name"`
		Price      float64               `form:"price"`
		FinalPrice float64               `form:"final_price"`
		Tags       string                `form:"tags"`
		Avatar     *multipart.FileHeader `form:"avatar"`
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
	if req.Avatar != nil {
		fs, err := req.Avatar.Open()
		if err != nil {
			RespBindError(c, err)
			return
		}
		defer fs.Close()

		img, err := io.ReadAll(fs)
		if err != nil {
			RespBindError(c, err)
			return
		}

		imageKey := storage.GetGoodsAvatarImageKey(req.ID)
		err = storage.SaveImage(imageKey, img)
		if err != nil {
			RespInternalError(c, err)
			return
		}
	}

	var now = time.Now()
	goods := storage.Goods{
		ID:         req.ID,
		LesseeID:   req.LesseeID,
		Name:       req.Name,
		Price:      req.Price,
		FinalPrice: req.FinalPrice,
		Tags:       GetTags(req.Tags),
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
