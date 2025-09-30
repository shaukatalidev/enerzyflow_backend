package users

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SaveProfileHandler(c *gin.Context) {
    var req SaveProfileRequest

    // Accept JSON directly
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }
    fmt.Println(req)

    userIDVal, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
        return
    }
    userID, ok := userIDVal.(uuid.UUID)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
        return
    }


    resp, err := SaveProfileService(userID.String(), req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "profile saved",
        "user":    resp.User,
        "company": resp.Company,
    })
}

func GetProfileHandler(c *gin.Context) {
    userIDVal, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
        return
    }
    userID, ok := userIDVal.(uuid.UUID)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
        return
    }

    resp, err := GetProfileService(userID.String())
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, resp)
}
