package handler

import (
	"mall/storage"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

func LesseeMiddle(c *gin.Context) {
	lid := c.Request.Header.Get("lessee")
	if lid != "" {
		id, err := strconv.Atoi(lid)
		if err != nil {
			logrus.Errorf("parse lessee id:%s error", lid)
		}
		c.Set("lid", uint64(id))
	}
}

func LocalSessionMiddle(c *gin.Context) {
	if strings.Contains(c.RemoteIP(), "127.0.0.1") {
		c.Set("uid", uint64(1))
	}
}

func GetSessionMiddle(admin bool, manage bool, jwtSecret string) func(c *gin.Context) {
	return func(c *gin.Context) {
		uid := c.GetInt("uid")
		if uid != 0 && admin {
			c.Next()
			return
		}

		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
			})
			return
		}

		// 验证token（这里简化处理）
		openID := verifyToken(token, jwtSecret)
		user, err := storage.Model[storage.User]().GetByOpenID(openID)
		if err != nil || user.ID == 0 {
			logrus.Errorf("get user by openid:%s error:%v", openID, err)
			RespUnauthorized(c)
			return
		}
		if admin && user.Kind != storage.Admin {
			RespUnauthorized(c)
			return
		}
		if manage {
			if user.Kind != storage.Admin && user.Kind != storage.Manger {
				RespUnauthorized(c)
				return
			}
		}
		c.Set("uid", user.ID)
	}
}

func verifyToken(tokenString, jwtSecret string) string {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	if v, ok := claims["openid"]; ok {
		openid, _ := v.(string)
		return openid
	}
	return ""
}
func generateJWTToken(openid, jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"openid": openid,
		"exp":    time.Now().Add(time.Hour * 24 * 7).Unix(), // 7天过期
		"iat":    time.Now().Unix(),
	})

	return token.SignedString([]byte(jwtSecret))
}
