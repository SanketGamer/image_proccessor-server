package main

import (
	"context"
	"image_service/config"
	"image_service/router"
	appSentry "image_service/sentry"
	"image_service/tasks"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
)

func main() {
	//load all config from .env
	cfg := config.Load()

	appSentry.Init(appSentry.Config{
		DSN:         cfg.SentryDSN,
		Environment: cfg.SentryEnv,
		Release:     cfg.SentryRelease,
		Debug:       cfg.SentryEnv == "development",
	})
	defer appSentry.Flush()

	//connect all service
	db := config.ConnectDB(cfg)
	rdb := config.ConnectRedis(cfg)
	s3Client := config.ConnectS3(cfg)

	// create Asynq client (enqueues jobs) and server (processes jobs), Both use Redis as the backbone
	redisopt := asynq.RedisClientOpt{Addr: cfg.RedisAddr}
	asynqClient := asynq.NewClient(redisopt)
	defer asynqClient.Close()

	// start the Asynq worker like same time 5 worker do simuntanously means 5 goroutine we trigger
	asynqServer := asynq.NewServer(redisopt, asynq.Config{
		Concurrency: 5, //worker can process 5 jobs simuntanously
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			appSentry.CaptureWorkerError(err, task.Type(), task.ResultWriter().TaskID())
		}),
	})

	processor := &tasks.ImageProcessor{
		DB:       db,
		S3Client: s3Client,
		S3Bucket: cfg.AWSBucket,
	}

	//gorila Mux is a router library in go
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeImageTransform, processor.ProcessImageTransform)

	//run worker in bg goroutine
	go func() {
		if err := asynqServer.Run(mux); err != nil {
			log.Fatalf("Asynq server error: %v", err)
		}
	}()
	//setup router
	r := router.Setup(router.Deps{
		DB:          db,
		Redis:       rdb,
		AsynqClient: asynqClient,
		Cfg:         cfg,
		S3Client:    s3Client,
	})
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}
	// Run in a goroutine so we can do graceful shutdown below
	go func() {
		log.Printf("Server running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	//shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	asynqServer.Shutdown()
	srv.Shutdown(ctx)
	log.Println("Server stopped cleanly")
}
