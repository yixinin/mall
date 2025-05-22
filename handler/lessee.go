package handler

import (
	"mall/set"
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
		ID     uint64 `uri:"id"`
		Name   string `json:"name"`
		Enable bool   `json:"enable"`
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

	user, err := storage.Model[storage.User]().GetByID(c.GetUint64("uid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	if user.Kind != storage.Admin {
		lessee, err := storage.Model[storage.Lessee]().GetByID(c.GetUint64("lessee"))
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if !set.From(lessee.Admins).Has(user.ID) {
			RespMessage(c, "not allow")
			return
		}
	}

	var status = storage.Enabled
	if !req.Enable {
		status = storage.Disabled
	}

	err = storage.Model[storage.Lessee]().Update(req.ID, req.Name, status)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, req.ID)
}

func (h *Handler) UpdateLesseeTech(c *gin.Context) {
	var req struct {
		ID  uint64   `uri:"id"`
		Add []uint64 `json:"add"`
		Del []uint64 `json:"del"`
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
	user, err := storage.Model[storage.User]().GetByID(c.GetUint64("uid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	if user.Kind != storage.Admin {
		lessee, err := storage.Model[storage.Lessee]().GetByID(c.GetUint64("lessee"))
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if !set.From(lessee.Admins).Has(user.ID) {
			RespMessage(c, "not allow")
			return
		}
	}
	err = storage.Model[storage.Lessee]().UpdateManger(req.ID, req.Add, req.Del)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, req.ID)
}

func (h *Handler) UpdateLesseeManager(c *gin.Context) {
	var req struct {
		ID  uint64   `uri:"id"`
		Add []uint64 `json:"add"`
		Del []uint64 `json:"del"`
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
	user, err := storage.Model[storage.User]().GetByID(c.GetUint64("uid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	if user.Kind != storage.Admin {
		lessee, err := storage.Model[storage.Lessee]().GetByID(c.GetUint64("lessee"))
		if err != nil {
			RespInternalError(c, err)
			return
		}
		if !set.From(lessee.Admins).Has(user.ID) {
			RespMessage(c, "not allow")
			return
		}
	}
	err = storage.Model[storage.Lessee]().UpdateTech(req.ID, req.Add, req.Del)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, req.ID)
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

func (h *Handler) GetLesseeMembers(c *gin.Context) {
	lessee, err := storage.Model[storage.Lessee]().GetByID(c.GetUint64("lid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	techs, err := storage.Model[storage.User]().GetUsersByIDs(lessee.Techs)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, techs)
}
func (h *Handler) GetLesseeAdmins(c *gin.Context) {
	lessee, err := storage.Model[storage.Lessee]().GetByID(c.GetUint64("lid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	admins, err := storage.Model[storage.User]().GetUsersByIDs(lessee.Admins)
	if err != nil {
		RespInternalError(c, err)
		return
	}

	Response(c, admins)
}
