package middleware

import (
	"net/http"
    "strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)


type Claims struct{
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}


func AuthMiddleware(jwtSecret string) gin.HandlerFunc{
	return func(ctx *gin.Context){
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == ""{
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error":"Authorization header required"})
			return
		}

		//split bearer token
		parts:=strings.SplitN(authHeader," ",2)
		if len(parts)!=2 || parts[0]!="Bearer"{
			ctx.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"error":"Invalid authorization format"})
			return
		}
		tokenStr:=parts[1]
		// Parse and verify the token using our secret key
        // jwt.ParseWithClaims decodes the token + checks signature + expiry
		claims:=&Claims{}
		token,err:=jwt.ParseWithClaims(tokenStr,claims,func(t *jwt.Token) (interface{},error){
			if _,ok:=t.Method.(*jwt.SigningMethodHMAC); !ok{
				return nil,jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret),nil
		})
		if err!=nil || !token.Valid{
			ctx.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"error": "Invalid or expired token",})
			return
		}
        // Store user ID in context so handlers can read it
        // ctx.Set is like passing a sticky note to the next function
		ctx.Set("userID",claims.UserID)
		ctx.Next()
	}
}