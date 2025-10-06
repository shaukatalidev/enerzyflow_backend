package routes

import (
	"enerzyflow_backend/internal/auth"
	"enerzyflow_backend/internal/orders"
	"enerzyflow_backend/internal/users"
	"enerzyflow_backend/utils"
	"github.com/gin-gonic/gin"
)

func RegisterAllRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Backend Running!")
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/send-otp", auth.SendOTPHandler)
		authGroup.POST("/verify-otp", auth.VerifyOTPHandler)
	}

	userGroup := r.Group("/users", utils.AuthMiddleware())
	{
		userGroup.POST("/profile", users.SaveProfileHandler)
		userGroup.GET("/profile", users.GetProfileHandler)
	}

	orderGroup := r.Group("/orders", utils.AuthMiddleware())
	{
		orderGroup.POST("/create", orders.CreateOrderHandler)
		orderGroup.GET("/get-all", orders.GetOrdersHandler)
		orderGroup.GET("/:id", orders.GetOrderHandler)
		orderGroup.POST("/:id/payment-screenshot",orders.UploadPaymentScreenshotHandler)
		orderGroup.PUT("/:id/status", utils.RoleMiddleware("printing"), orders.UpdateOrderStatusHandler)
		orderGroup.GET("/get-all-orders", utils.RoleMiddleware("printing"), orders.GetAllOrdersForPrinting)
		orderGroup.GET("/:id/tracking", orders.GetOrderTrackingHandler)
	}
}

// func RegisterAuthRoutes(r *gin.Engine) {
// 	authGroup := r.Group("/auth")
// 	authGroup.POST("/send-otp", auth.SendOTPHandler)
// 	authGroup.POST("/verify-otp", auth.VerifyOTPHandler)
// }

// func RegisterUserRoutes(r *gin.Engine) {
// 	userGroup := r.Group("/users", utils.AuthMiddleware())
// 	userGroup.POST("/profile", users.SaveProfileHandler)
// 	userGroup.GET("/profile", users.GetProfileHandler)
// }

// func RegisterOrderRoutes(r *gin.Engine) {
// 	orderGroup := r.Group("/orders", utils.AuthMiddleware())
// 	orderGroup.POST("/", orders.CreateOrderHandler)
// 	orderGroup.GET("/", orders.GetOrdersHandler)
// 	orderGroup.GET("/:id", orders.GetOrderHandler)
// 	orderGroup.PUT("/:id/status", orders.UpdateOrderStatusHandler)
// 	orderGroup.DELETE("/:id", orders.DeleteOrderHandler)
// }
