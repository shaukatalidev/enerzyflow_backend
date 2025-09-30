package auth

import (
	"fmt"
	"net/http"

	"enerzyflow_backend/internal/users"
	"enerzyflow_backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func SendOTPHandler(c *gin.Context) {
	var req SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var role string
	u, err := users.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if u != nil {
		role = u.Role
	} else {
		role = "business_owner"
	}

	otp, err := SendOTP(req.Email, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully", "otp": otp})
}

func VerifyOTPHandler(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	u, err := users.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	var role string
	if u != nil {
		role = u.Role
	} else {
		role = "business_owner"
	}
	fmt.Println("Role:", role)
	valid, expired, err := VerifyOTP(req.Email, role, req.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "verification failed"})
		return
	}
	if !valid {
		if expired {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OTP expired"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid OTP"})
		return
	}

	if u == nil {
		newUser := &users.User{
			UserID: uuid.New().String(),
			Email:  req.Email,
			Role:   role,
		}
		if err := users.InsertUser(newUser); err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23514" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role value: role must be one of the allowed values"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
		u = newUser
	}

	userUUID, err := uuid.Parse(u.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	accessToken, err := utils.GenerateTokens(u.Email, userUUID, u.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "OTP verified successfully",
		"accessToken": accessToken,
		"user": gin.H{
			"user_id":     u.UserID,
			"email":       u.Email,
			"name":        u.Name,
			"phone":       u.Phone,
			"designation": u.Designation,
			"role":        u.Role,
		},
	})
}