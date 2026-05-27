package handlers

import (
	"encoding/json"
	"fmt"
	"image_service/services"
	appSentry "image_service/sentry"
	"image_service/tasks"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)


type ImageHandler struct {
    ImageService *services.ImageService
    AsynqClient  *asynq.Client
    S3Bucket     string
}

func (h *ImageHandler) Upload(ctx *gin.Context){
    userID:=ctx.GetString("userID")
    // FormFile reads the uploaded file from the form field named "image"
    file,header,err:=ctx.Request.FormFile("image")
     if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
        return
    }
    defer file.Close()
    //limit file size to 10mb
    if header.Size > 10<<20{
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 10MB)"})
        return
    }
    img,err:=h.ImageService.UploadImage(file,header,userID)
    if err!=nil{
        appSentry.CaptureGinError(ctx, err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusCreated, gin.H{"image": img})
}

// Transform handles POST /images/:id/transform
// It enqueues a background job — doesn't block the request
func (h *ImageHandler) Transform(ctx *gin.Context){
    userID:= ctx.GetString("userID")
    imageID:= ctx.Param("id")
    //fetch the image to get its s3 key
    img,err:=h.ImageService.GetImage(imageID,userID)
    if err!=nil{
        appSentry.CaptureGinError(ctx, err)
        ctx.JSON(http.StatusNotFound,gin.H{"error":"image not found"})
        return
    }
     // Parse transformation options from request body
     var payload tasks.TransformPayload
     if err:=ctx.ShouldBindJSON(&payload); err!=nil{
        appSentry.CaptureGinError(ctx, err)
        ctx.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
        return
     }
     //fill the required fields
     payload.ImageID=imageID
     payload.UserID=userID
     payload.S3Key=extractS3Key(img.OriginalURL)
     payload.OutputKey=fmt.Sprintf("transformed/%s/%s", userID, uuid.NewString()+".jpg")

     //serilize payload to json for asynq
     payloadBytes,_:=json.Marshal(payload)
     

    // Create and enqueue the Asynq task
    // asynq.NewTask(type, payload) = create the "order ticket"
    // asynqClient.Enqueue(task) = put it in the kitchen queue
    task:=asynq.NewTask(tasks.TypeImageTransform,payloadBytes)
    info,err:=h.AsynqClient.Enqueue(task)
    if err!=nil{
       appSentry.CaptureGinError(ctx, err)
       ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue task"})
       return
    }
    // Mark image as "processing" in DB
    h.ImageService.UpdateStatus(imageID,"processing")
    ctx.JSON(http.StatusAccepted,gin.H{
        "message": "transformation queued",
        "task_id": info.ID,
        "image":   img,
    })
}

// GetImage handles GET /images/:id
func (h *ImageHandler) GetImage(ctx *gin.Context){
    userID:=ctx.GetString("userID")
    imageID:=ctx.Param("id")
    img,err:=h.ImageService.GetImage(imageID,userID)
    if err!=nil{
        appSentry.CaptureGinError(ctx, err)
        ctx.JSON(http.StatusNotFound,gin.H{"error":"image not found"})
        return
    }
    ctx.JSON(http.StatusOK,gin.H{"image":img})
}

// ListImages handles GET /images?page=1&limit=10
func (h *ImageHandler) ListImages(ctx *gin.Context){
    userID:=ctx.GetString("userID")

    page,_:=strconv.Atoi(ctx.DefaultQuery("page","1"))
    limit,_:=strconv.Atoi(ctx.DefaultQuery("limit","10"))
    // Clamp limit to prevent huge queries
    if limit>50 {limit=50}
    if page<1 {page=1}

    images,total,err:=h.ImageService.ListImage(userID,page,limit)
    if err != nil {
        appSentry.CaptureGinError(ctx, err)
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK,gin.H{
        "images":images,
        "meta":gin.H{
            "total":total,
            "page":page,
            "limit":limit,
            "pages": (total + int64(limit) - 1) / int64(limit),
        },
    })
}

// extractS3Key gets the S3 key from a full URL, ex: "https://bucket.s3.amazonaws.com/images/abc.jpg" → "images/abc.jpg"
func extractS3Key(url string) string {
    for i:=len(url)-1;i>=0;i--{
        if url[i]=='/' && i>10{
         // Find third slash onwards
        }
    }
        //Simple split approach
        parts:=splitAfterNth(url,"/",3)
        if len(parts)>1{
            return parts[1]
        }
        return url
}

func splitAfterNth(s,sep string,n int) []string{
    idx:=0
    for i:=0;i<n;i++{
        pos:=indexOf(s[idx:],sep)
        if pos<0{ return []string{s}}
        idx+=pos+len(sep)
    }
    return []string{s[:idx-1], s[idx:]}
}

func indexOf(s, sub string) int {
    for i := 0; i <= len(s)-len(sub); i++ {
        if s[i:i+len(sub)] == sub { return i }
    }
    return -1
}
