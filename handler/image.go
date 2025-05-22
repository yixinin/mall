package handler

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mall/storage"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/disintegration/imaging"
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

func (h *Handler) PostImage(c *gin.Context) {
	idStr := c.Param("id")
	var id uint64
	if idStr != "" {
		id, _ = strconv.ParseUint(idStr, 10, 64)
	}
	fh, err := c.FormFile("avatar")
	if err != nil {
		RespInternalError(c, err)
		return
	}
	f, err := fh.Open()
	if err != nil {
		RespInternalError(c, err)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	if id == 0 {
		id, err = storage.GenID()
		if err != nil {
			RespInternalError(c, err)
			return
		}
	}

	key := storage.GetGoodsAvatarImageKey(id)
	data, err = compressAndCropSquare(data, 80, 150, path.Ext(fh.Filename))
	if err != nil {
		logrus.Errorf("compress image error:%v", err)
		RespInternalError(c, err)
		return
	}
	err = storage.SaveImage(key, data)
	if err != nil {
		RespInternalError(c, err)
		return
	}
	var ack struct {
		ID   uint64 `json:"id"`
		Path string `json:"path"`
	}
	ack.ID = id
	ack.Path = fmt.Sprintf("/%s", key)
	Response(c, ack)
}

func compressAndCropSquare(data []byte, quality int, toSize int, ext string) ([]byte, error) {
	opts := imaging.AutoOrientation(true)
	src, err := imaging.Decode(bytes.NewReader(data), opts)
	if err != nil {
		return nil, err
	}
	// 获取原始图片尺寸
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 计算正方形裁剪区域
	var cropRect image.Rectangle
	if width > height {
		// 横向图片，从中心裁剪
		cropX := (width - height) / 2
		cropRect = image.Rect(cropX, 0, cropX+height, height)
	} else {
		// 纵向图片或正方形，从中心裁剪
		cropY := (height - width) / 2
		cropRect = image.Rect(0, cropY, width, cropY+width)
	}

	// 裁剪图片
	croppedImg := imaging.Crop(src, cropRect)
	croppedImg = imaging.Resize(croppedImg, toSize, toSize, imaging.Lanczos)

	var w bytes.Buffer
	// 根据格式保存图片
	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&w, croppedImg, &jpeg.Options{Quality: quality})
	case ".png":
		err = png.Encode(&w, croppedImg)
	default:
		// 默认使用JPEG格式
		err = jpeg.Encode(&w, croppedImg, &jpeg.Options{Quality: quality})
	}
	return w.Bytes(), err
}
