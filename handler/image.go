package handler

import (
	"bytes"
	"errors"
	"io"
	"mall/storage"
	"net/http"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetImage(c *gin.Context) {
	key := c.Request.URL.Path
	key = strings.TrimPrefix(key, "/")
	var data []byte
	err := storage.GetDB().View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		data, err = item.ValueCopy(nil)
		return err
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		c.Status(404)
		c.Abort()
		return
	}

	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, bytes.NewReader(data))
	if err != nil {
		logrus.Errorf("get image error:%v", err)
	}
}
