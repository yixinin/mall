package handler

import (
	"mall/storage"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PostLessee(c *gin.Context) {
	var req struct {
		Name   string `json:"name"`
		Enable bool   `json:"enable"`
	}
	err := c.BindJSON(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	id, err := storage.GenID()
	if err != nil {
		RespInternalError(c, err)
		return
	}
	var now = time.Now()
	var lessee = &storage.Lessee{
		ID:         id,
		Name:       req.Name,
		Status:     storage.Enabled,
		CreateTime: now,
		UpdateTime: now,
	}
	if req.Enable {
		lessee.Status = storage.Enabled
	} else {
		lessee.Status = storage.Disabled
	}

	err = lessee.Save()
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, id)
}

func (h *Handler) GetLessee(c *gin.Context) {
	var req struct {
		ID uint64 `uri:"id"`
	}
	err := c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	lessee, err := storage.Model[storage.Lessee]().GetByID(req.ID)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, lessee)
}

func (h *Handler) GetLesseeList(c *gin.Context) {
	ls, err := storage.Model[storage.Lessee]().GetLessees()
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, ls)
}

func (h *Handler) PutLessee(c *gin.Context) {
	var req struct {
		ID        uint64   `uri:"id"`
		Name      string   `json:"name"`
		Enable    bool     `json:"enable"`
		AddAdmins []uint64 `json:"add_admins"`
		DelAdmins []uint64 `json:"del_admins"`
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
	var status = storage.Enabled
	if !req.Enable {
		status = storage.Disabled
	}

	err = storage.Model[storage.Lessee]().Update(req.ID, req.Name, status, req.AddAdmins, req.DelAdmins)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, "")
}

func (h *Handler) DeleteLessee(c *gin.Context) {
	var req struct {
		ID uint64 `uri:"id"`
	}
	err := c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	err = storage.Model[storage.Lessee]().Delete(req.ID)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, "")
}
