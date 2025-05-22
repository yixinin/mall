package handler

import (
	"mall/set"
	"mall/storage"
	"net/http"
	"strconv"
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

func GetSessionMiddle(jwtSecret string) func(c *gin.Context) {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
			})
			return
		}

		openID := verifyToken(token, jwtSecret)
		user, err := storage.Model[storage.User]().GetByOpenID(openID)
		if err != nil || user.ID == 0 {
			logrus.Errorf("get user by openid:%s error:%v", openID, err)
			RespUnauthorized(c)
			return
		}

		c.Set("uid", user.ID)
		c.Set("role", string(user.Kind))
		c.Set("user", user)
	}
}

func RoleMiddle(roles ...storage.UserKind) func(c *gin.Context) {
	roleSet := set.From(roles)
	return func(c *gin.Context) {
		role := storage.UserKind(c.GetString("role"))
		if !roleSet.Has(role) {
			logrus.Infof("role:%s, forbidden", role)
			RespForbidden(c)
			c.Abort()
			return
		}
		lessee, err := storage.Model[storage.Lessee]().GetByID(c.GetUint64("lid"))
		if err != nil {
			RespForbidden(c)
			c.Abort()
			logrus.Infof("lessee:%d not found, forbidden", c.GetUint64("lid"))
			return
		}
		switch role {
		case storage.Manger:
			if !set.From(lessee.Admins).Has(c.GetUint64("uid")) {
				RespForbidden(c)
				c.Abort()
				logrus.Infof("user:%d is not manager, forbidden", c.GetUint64("uid"))
				return
			}
		case storage.Technician:
			if !set.From(lessee.Techs).Has(c.GetUint64("uid")) {
				RespForbidden(c)
				c.Abort()
				logrus.Infof("user:%d is not tech, forbidden", c.GetUint64("uid"))
				return
			}
		}
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
		"exp":    time.Now().AddDate(2000, 0, 0).Unix(), // 不过期
		"iat":    time.Now().Unix(),
	})

	return token.SignedString([]byte(jwtSecret))
}
