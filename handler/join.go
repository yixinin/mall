package handler

import (
	"mall/storage"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (h Handler) PostJoin(c *gin.Context) {
	lessee, err := storage.Model[storage.Lessee]().GetByID(c.GetUint64("lid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	user, err := storage.Model[storage.User]().GetByID(c.GetUint64("uid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	id, err := storage.GenID()
	if err != nil {
		RespInternalError(c, err)
		return
	}
	var now = time.Now()
	join := storage.Join{
		ID: id,
		User: storage.SimpleUser{
			ID:       user.ID,
			Nickname: user.Nickname,
		},
		Lessee: storage.IdName{
			Id:   lessee.ID,
			Name: lessee.Name,
		},
		CreateTime: now,
		UpdateTime: now,
	}
	err = join.Save()
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, id)
}

func (h Handler) GetJoins(c *gin.Context) {
	joins, err := storage.Model[storage.Join]().GetJoins(c.GetUint64("lid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, joins)
}

func (h Handler) PutJoin(c *gin.Context) {
	var req struct {
		ID     uint64             `uri:"id"`
		Status storage.JoinStatus `json:"status"`
	}
	err := c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	err = c.Bind(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}

	err = storage.Model[storage.Join]().Update(c.GetUint64("lid"), req.ID, req.Status)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, req.ID)
}
func (h *Handler) DeleteJoin(c *gin.Context) {
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
	err = storage.Model[storage.Join]().Delete(lid, id)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, id)
}
