package utils

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const Secretkey = "totallsecretkeylol"

func GenerateTokens(email string, userid uuid.UUID, role string) (string, error) {
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "Email":  email,
        "userid": userid.String(),
        "role":   role,
        "exp":    time.Now().Add(24 * time.Hour).Unix(),
    })
    access, err := accessToken.SignedString([]byte(Secretkey))
    if err != nil {
        return "",  err
    }

    return access, nil
}

func VerifyToken(token string) (error, uuid.UUID, bool, string) {
    tokenparsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(Secretkey), nil
    }, jwt.WithLeeway(5*time.Second), jwt.WithValidMethods([]string{"HS256"}))

    if err != nil {
        if errors.Is(err, jwt.ErrTokenExpired) {
            return nil, uuid.Nil,  true, ""
        }
        return err, uuid.Nil,  false, ""
    }

    claims, ok := tokenparsed.Claims.(jwt.MapClaims)
    if !ok || !tokenparsed.Valid {
        return errors.New("invalid token claims"), uuid.Nil,  false, ""
    }

    useridStr, ok := claims["userid"].(string)
    if !ok {
        return errors.New("userid not found or invalid"), uuid.Nil,  false, ""
    }
    userid, err := uuid.Parse(useridStr)
    if err != nil {
        return errors.New("invalid userid format"), uuid.Nil,  false, ""
    }

    // // roleidStr, ok := claims["roleid"].(string)
    // if !ok {
    //     return errors.New("roleid not found or invalid"), uuid.Nil,false, ""
    // }
    // // roleid, err := uuid.Parse(roleidStr)
    // if err != nil {
    //     return errors.New("invalid roleid format"), uuid.Nil,  false, ""
    // }

    role, ok := claims["role"].(string)
    if !ok {
        return errors.New("role not found in token"), uuid.Nil, false, ""
    }

    expRaw, ok := claims["exp"].(float64)
    if !ok {
        return errors.New("invalid exp in token"), uuid.Nil, false, ""
    }
    expired := time.Now().Unix() > int64(expRaw)

    return nil, userid, expired, role
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

        err, userID, expired, role := VerifyToken(tokenStr)
        // _ = roleId
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
            return
        }
        if expired {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Access token expired"})
            return
        }
        c.Set("role", role)
        c.Set("user_id", userID)
        c.Next()
    }
}

func ExtractClaimsWithoutValidation(tokenStr string) (jwt.MapClaims, error) {
    parsedToken, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
    if err != nil {
        return nil, err
    }

    claims, ok := parsedToken.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }
    return claims, nil
}

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "role not found in context"})
			return
		}

		role, ok := roleVal.(string)
		if !ok || role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid role in context"})
			return
		}

		if role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied for this role"})
			return
		}

		c.Next()
	}
}