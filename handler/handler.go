package handler

import (
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

	noauth := e.Group("/api/v1/local")

	noauth.PUT("/user/:id", h.PutUser)
	noauth.GET("/user", h.GetUsers)
	noauth.POST("/goods", h.PostGoods)
	noauth.PUT("/goods/:id", h.PutGoods)
	noauth.GET("/goods", h.GetGoodsList)
	noauth.GET("/goods/id", h.GetGoods)
	noauth.GET("/order", h.GetOrders)
	noauth.PUT("/order/:id", h.PutOrder)
	noauth.POST("/lessee", h.PostLessee)
	noauth.PUT("/lessee/:id", h.PutLessee)
	noauth.GET("/lessee", h.GetLesseeList)
	noauth.DELETE("/lessee/:id", h.DeleteLessee)

	api := e.Group("/api/v1/mini")
	admin := e.Group("/api/v1/admin", LocalSessionMiddle, GetSessionMiddle(true, false, h.jwtSecret))
	manage := e.Group("/api/v1/manage/mini", GetSessionMiddle(true, false, h.jwtSecret))

	goods := api.Group("/goods")
	goods.GET("/", h.GetGoodsList)
	goods.GET("/id", h.GetGoods)

	order := api.Group("/order", GetSessionMiddle(false, false, h.jwtSecret))
	order.GET("/", h.GetOrders)
	order.POST("/", h.PostOrder)
	order.PUT("/order/:id", h.PutOrder)

	user := api.Group("/user")
	user.GET("/info", h.PreLogin)
	user.PUT("/:id", h.PutUser)
	user.GET("/:id", h.GetUser)

	manage.GET("/goods/:id", h.GetGoods)
	manage.GET("/goods", h.GetGoodsList)
	manage.POST("/goods", h.PostGoods)
	manage.PUT("/goods/:id", h.PutGoods)
	manage.DELETE("/goods/:id", h.DeleteGoods)
	manage.GET("/order", h.GetOrders)
	manage.POST("/order", h.PostOrder)
	manage.PUT("/order/:id", h.PutOrder)
	manage.GET("/user/:id", h.GetUser)
	manage.GET("/user", h.GetUsers)
	manage.PUT("/user/:id", h.PutUser)

	admin.GET("/user/:id", h.GetUser)
	admin.GET("/user", h.GetUsers)
	admin.PUT("/user/:id", h.PutUser)
	admin.DELETE("/user/:id", h.DeleteUser)

	admin.GET("/goods/:id", h.GetGoods)
	admin.GET("/goods", h.GetGoodsList)
	admin.POST("/goods", h.PostGoods)
	admin.PUT("/goods/:id", h.PutGoods)
	admin.DELETE("/goods/:id", h.DeleteGoods)

	admin.GET("/order", h.GetOrders)
	admin.GET("/order/:id", h.GetOrders)
	admin.POST("/order", h.PostOrder)
	admin.PUT("/order/:id", h.PutOrder)
	admin.DELETE("/order/:id", h.DeleteOrder)

	admin.POST("/lessee", h.PostLessee)
	admin.PUT("/lessee/:id", h.PutLessee)
	admin.GET("/lessee", h.GetLesseeList)
	admin.GET("/lessee/:id", h.GetLessee)
	admin.DELETE("/lessee/:id", h.DeleteLessee)

}
