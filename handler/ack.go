package handler

import (
	"errors"
	"net/http"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Ack[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    T      `json:"data"`
}

func Response[T any](c *gin.Context, data T) {
	ack := Ack[T]{
		Data: data,
	}
	c.JSON(http.StatusOK, ack)
}

func RespUnauthorized(c *gin.Context) {
	c.Status(http.StatusUnauthorized)
	c.Abort()
}

func RespInternalError(c *gin.Context, err error) {
	var ack Ack[any]
	if errors.Is(err, badger.ErrKeyNotFound) {
		ack = Ack[any]{
			Code:    404,
			Message: "resource not foud",
		}
		logrus.WithField("path", c.Request.URL.Path).Errorf("resource not foud")
	} else {
		ack = Ack[any]{
			Code:    500,
			Message: "interval error",
		}
		logrus.WithField("path", c.Request.URL.Path).Errorf("handle with interval error:%v", err)
	}

	c.JSON(http.StatusOK, ack)
}

func RespBindError(c *gin.Context, err error) {
	if err != nil {
		logrus.WithField("path", c.Request.URL.Path).Errorf("bind error:%v", err)
	}
	ack := Ack[any]{
		Code:    400,
		Message: "params error",
	}
	c.JSON(http.StatusOK, ack)
}

func RespMessage(c *gin.Context, msg string) {
	if msg == "" {
		msg = "params error"
	}
	ack := Ack[any]{
		Code:    400,
		Message: msg,
	}
	c.JSON(http.StatusOK, ack)
}
