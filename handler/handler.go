package handler

import (
	"mall/storage"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	wxApp     *miniProgram.MiniProgram
	jwtSecret string
}

func NewHandler(e *gin.Engine, appid, secret, jwtSecret string) *Handler {
	miniProgram, err := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:  appid,
		Secret: secret,
		OAuth:  miniProgram.OAuth{
			// Callback: "YOUR_CALLBACK_URL", // 可选，根据业务需要
		},
	})

	if err != nil {
		panic(err)
	}

	h := &Handler{
		jwtSecret: jwtSecret,
		wxApp:     miniProgram,
	}

	h.Register(e)

	return h
}

func (h *Handler) Register(e *gin.Engine) {
	e.GET("/api/health", func(c *gin.Context) {
		c.String(200, "alive")
	})
	e.POST("/api/v1/login", h.Login)
	e.GET("/img/:target/:type/:id", h.GetImage)

	e.Use(LesseeMiddle)

	api := e.Group("/api/v1/mini")

	api.POST("/image", h.PostImage)
	api.POST("/image/:id", h.PostImage)

	goods := api.Group("/goods")
	goods.GET("", h.GetGoodsList)
	goods.GET("/pre", h.PreGetGoodsList)
	goods.GET("/manage", GetSessionMiddle(h.jwtSecret), h.GetGoodsList)
	goods.GET("/manage/pre", GetSessionMiddle(h.jwtSecret), h.PreGetGoodsList)
	goods.GET("/:id", h.GetGoods)
	goods.HEAD("/:id", h.GetGoods)
	goods.POST("", GetSessionMiddle(h.jwtSecret), h.PostGoods)
	goods.PUT("/:id", GetSessionMiddle(h.jwtSecret), h.PutGoods)
	goods.DELETE("/:id", GetSessionMiddle(h.jwtSecret), h.DeleteGoods)

	order := api.Group("/order", GetSessionMiddle(h.jwtSecret))
	order.GET("", h.GetOrders)
	order.GET("/pre", h.PreGetOrders)
	order.GET("/:id", h.GetOrder)
	order.HEAD("/:id", h.GetOrder)
	order.POST("", h.PostOrder)
	order.PUT("/:id", h.PutOrder)
	order.DELETE("/:id", RoleMiddle(storage.Admin), h.DeleteOrder)

	user := api.Group("/user", GetSessionMiddle(h.jwtSecret))
	user.GET("/info", h.PreLogin)
	user.HEAD("/info", h.PreLogin)
	user.PUT("/:id", h.PutUser)
	user.GET("/:id", h.GetUser)
	user.HEAD("/:id", h.GetUser)
	user.GET("", RoleMiddle(storage.Admin, storage.Manger), h.GetUsers)
	user.DELETE("/:id", h.DeleteUser)

	lessee := api.Group("/lessee", GetSessionMiddle(h.jwtSecret))
	lessee.POST("", RoleMiddle(storage.Admin), h.PostLessee)
	lessee.PUT("/:id", RoleMiddle(storage.Admin), h.PutLessee)
	lessee.PUT("/:id/manager", RoleMiddle(storage.Admin, storage.Manger), h.UpdateLesseeManager)
	lessee.PUT("/:id/tech", RoleMiddle(storage.Admin, storage.Manger), h.UpdateLesseeTech)
	lessee.GET("", h.GetLesseeList)
	lessee.GET("/:id", h.GetLessee)
	lessee.DELETE("/:id", RoleMiddle(storage.Admin), h.DeleteLessee)
	lessee.GET("/:id/tech", RoleMiddle(storage.Admin, storage.Manger), h.GetLesseeMembers)
	lessee.GET("/:id/manager", RoleMiddle(storage.Admin), h.GetLesseeAdmins)

	join := api.Group("/join", GetSessionMiddle(h.jwtSecret))
	join.POST("", h.PostJoin)
	join.GET("", RoleMiddle(storage.Admin, storage.Manger), h.GetJoins)
	join.PUT("/:id", RoleMiddle(storage.Admin, storage.Manger), h.PutJoin)
	join.DELETE("/:id", RoleMiddle(storage.Admin, storage.Manger), h.DeleteJoin)

}
