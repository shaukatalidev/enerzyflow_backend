package users

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

func SaveProfileHandler(c *gin.Context) {
    var req SaveProfileRequest
    payload := c.PostForm("payload")
    if payload == "" {
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
            return
        }
    } else {
        if err := json.Unmarshal([]byte(payload), &req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload json"})
            return
        }
    }

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

    _ = os.MkdirAll("uploads/profiles", 0755)
    _ = os.MkdirAll("uploads/company_logos", 0755)
    _ = os.MkdirAll("uploads/labels", 0755)

    if file, err := c.FormFile("profile_img"); err == nil && file != nil {
        ext := strings.ToLower(filepath.Ext(file.Filename))
        filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
        dst := filepath.Join("uploads", "profiles", filename)
        if err := c.SaveUploadedFile(file, dst); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save profile image"})
            return
        }
        req.Profile.ProfileURL = "/uploads/profiles/" + filename
    }

    if file, err := c.FormFile("company_logo"); err == nil && file != nil {
        ext := strings.ToLower(filepath.Ext(file.Filename))
        filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
        dst := filepath.Join("uploads", "company_logos", filename)
        if err := c.SaveUploadedFile(file, dst); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save company logo"})
            return
        }
        req.Company.Logo = "/uploads/company_logos/" + filename
    }

    for i := 0; i < 50; i++ {
        key := fmt.Sprintf("label_%d", i)
        file, err := c.FormFile(key)
        if err != nil || file == nil {
            break
        }

        ext := strings.ToLower(filepath.Ext(file.Filename))
        filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
        dst := filepath.Join("uploads", "labels", filename)
        if err := c.SaveUploadedFile(file, dst); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save label image"})
            return
        }

        for len(req.Company.Labels) <= i {
            req.Company.Labels = append(req.Company.Labels, struct {
                LabelID string `json:"label_id"`
                Name    string `json:"name"`
                URL     string `json:"url"`
            }{})
        }

        req.Company.Labels[i].URL = "/uploads/labels/" + filename
        if name := c.PostForm(fmt.Sprintf("label_name_%d", i)); name != "" {
            req.Company.Labels[i].Name = name
        } else if req.Company.Labels[i].Name == "" {
            base := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
            req.Company.Labels[i].Name = base
        }
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
