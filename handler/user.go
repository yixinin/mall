package handler

import (
	"errors"
	"mall/storage"
	"net/http"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
)

func (h *Handler) PreLogin(c *gin.Context) {
	user, err := storage.Model[storage.User]().GetByID(c.GetUint64("uid"))
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, user)
}

func (h *Handler) Login(c *gin.Context) {
	// 1. 解析请求参数
	var req struct {
		Code      string `json:"code" form:"code"`
		Nickname  string `json:"nickname,omitempty" form:"nickname"`
		AvatarURL string `json:"avatar_url,omitempty" form:"avatar_url"`
	}

	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// 2. 使用SDK获取session信息
	session, err := h.wxApp.Auth.Session(c.Request.Context(), req.Code)
	if err != nil {
		RespInternalError(c, err)
		return
	}

	user, err := storage.Model[storage.User]().GetByOpenID(session.OpenID)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		RespInternalError(c, err)
		return
	}

	var now = time.Now()
	if errors.Is(err, badger.ErrKeyNotFound) {
		id, err := storage.GenID()
		if err != nil {
			RespInternalError(c, err)
			return
		}
		user = storage.User{
			ID:         id,
			OpenID:     session.OpenID,
			Avatar:     req.AvatarURL,
			Nickname:   req.Nickname,
			CreateTime: now,
			UpdateTime: now,
		}
		if user.OpenID == "oy5Gt7cgoPLhBcKSrickAfnREA2A" {
			user.Kind = storage.Admin
		} else {
			user.Kind = storage.Customer
		}
		err = user.Save()
		if err != nil {
			RespInternalError(c, err)
			return
		}
	} else {
		user.Avatar = req.AvatarURL
		user.Nickname = req.Nickname
		user.UpdateTime = now
		err := user.Update(user.ID)
		if err != nil {
			RespInternalError(c, err)
			return
		}
	}

	token, err := generateJWTToken(session.OpenID, h.jwtSecret)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	// 4. 返回响应
	c.JSON(http.StatusOK, gin.H{
		"openid": session.OpenID,
		// "session_key": session.SessionKey, // 注意：实际生产环境不应返回给前端
		"token":    token,
		"userinfo": user,
	})
}

func (h *Handler) PutUser(c *gin.Context) {
	var req struct {
		ID       uint64 `uri:"id"`
		Kind     string `json:"kind"`
		Avatar   string `json:"avatar"`
		Nickname string `json:"nickname"`
	}
	err := c.BindJSON(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	var user = &storage.User{
		Kind:     storage.UserKind(req.Kind),
		Avatar:   req.Avatar,
		Nickname: req.Nickname,
	}

	err = user.Update(req.ID)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, user)
}

func (h *Handler) GetUsers(c *gin.Context) {
	users, err := storage.Model[storage.User]().GetUsers()
	if err != nil {
		RespInternalError(c, err)
		return
	}
	Response(c, users)
}

func (h *Handler) GetUser(c *gin.Context) {
	var req struct {
		ID uint64 `uri:"id"`
	}
	err := c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	user, err := storage.Model[storage.User]().GetByID(req.ID)
	if errors.Is(err, badger.ErrKeyNotFound) {
		RespBindError(c, err)
		return
	}
	if err != nil {
		RespInternalError(c, err)
		return
	}

	Response(c, user)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	var req struct {
		ID uint64 `uri:"id"`
	}
	err := c.BindUri(&req)
	if err != nil {
		RespBindError(c, err)
		return
	}
	err = storage.Model[storage.User]().Delete(req.ID)
	if err != nil {
		RespInternalError(c, err)
		return
	}

	Response(c, "")
}
