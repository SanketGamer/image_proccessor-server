package router

import (
	"image_service/config"
	"image_service/handlers"
	"image_service/middleware"
	appSentry "image_service/sentry"
	"image_service/services"
	"time"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Deps struct {
	DB          *gorm.DB
	Redis       *redis.Client
	AsynqClient *asynq.Client
	Cfg         *config.Config
	S3Client    interface{} 
}

func Setup(deps Deps) *gin.Engine {
	r := gin.New()
	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(appSentry.Middleware())

	//init service
	authSvc := &services.AuthService{
		DB:        deps.DB,
		JWTSecret: deps.Cfg.JWTSecret,
		JWTExpiry: deps.Cfg.JWTExpiryHours,
	}
	imgSvc := &services.ImageService{
		DB:       deps.DB,
		Redis:    deps.Redis,
		S3Bucket: deps.Cfg.AWSBucket,
		S3Client: deps.S3Client.(*s3.Client),
	}
	//init handler
	authH := &handlers.AuthHandler{AuthService: authSvc}
	imgH := &handlers.ImageHandler{
		ImageService: imgSvc,
		AsynqClient:  deps.AsynqClient,
		S3Bucket:     deps.Cfg.AWSBucket,
	}
	// Public routes — no JWT needed
	public := r.Group("/api/v1")
	{
		public.POST("/register", authH.Register)
		public.POST("/login", authH.Login)
	}
	// Protected routes — JWT required
	// RateLimiter: max 30 requests per minute per user
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(deps.Cfg.JWTSecret))
	protected.Use(middleware.RateLimiter(deps.Redis, 30, time.Minute))
	{
		protected.POST("/images", imgH.Upload)
		protected.GET("/images", imgH.ListImages)
		protected.GET("/images/:id", imgH.GetImage)
		protected.POST("/images/:id/transform", imgH.Transform)
	}
	return r
}
