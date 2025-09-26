package routes

import (
    "enerzyflow_backend/internal/auth"
    "enerzyflow_backend/internal/users"
    "enerzyflow_backend/utils"
    "github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine) {
    api := r.Group("/auth")
    api.POST("/send-otp", auth.SendOTPHandler)
    api.POST("/verify-otp", auth.VerifyOTPHandler)
}

func RegisterUserRoutes(r *gin.Engine) {
    api := r.Group("/users", utils.AuthMiddleware())
    api.POST("/profile", users.SaveProfileHandler)
    api.GET("/profile", users.GetProfileHandler)
}
