package handlers

import (
	"image_service/services"
	"net/http"
	appSentry "image_service/sentry"
	"github.com/gin-gonic/gin"
)


type AuthHandler struct{
	AuthService *services.AuthService
}

type registerRequest struct{
	Username string `json:"username" binding:"required,min=3"`
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=5"`
}

func (h *AuthHandler) Register(ctx *gin.Context){
	var req registerRequest
	//reads json, convert json into go, stores values inside struct
	if err:=ctx.ShouldBindJSON(&req); err!=nil{
        appSentry.CaptureGinError(ctx, err)
		ctx.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return
	}
	user,token,err:=h.AuthService.Register(req.Username,req.Email,req.Password)
	if err!=nil{
            appSentry.CaptureGinError(ctx, err)
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
        return
	}
	ctx.JSON(http.StatusCreated,gin.H{"user":user,"token":token})
}

func (h *AuthHandler) Login(ctx *gin.Context){
	var req struct {
        Email    string `json:"email"    binding:"required,email"`
        Password string `json:"password" binding:"required"`
    }
	if err := ctx.ShouldBindJSON(&req); err != nil {
          appSentry.CaptureGinError(ctx, err)
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
	user,token,err:=h.AuthService.Login(req.Email,req.Password)
	if err!=nil{
           appSentry.CaptureGinError(ctx, err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
	}
	ctx.JSON(http.StatusOK,gin.H{"user":user,"token":token})
}