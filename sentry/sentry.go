package sentry

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gin-gonic/gin"
)

type Config struct {
	DSN         string
	Environment string
	Release     string
	Debug       bool
}

//initialize sentry
func Init(cfg Config) {
	if cfg.DSN == "" {
		log.Println("Sentry DSN not set — error tracking disabled")
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		Debug:            cfg.Debug, 
		TracesSampleRate: 1.0,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			if event.Request != nil && event.Request.Headers != nil {
				delete(event.Request.Headers, "Authorization")
				delete(event.Request.Headers, "Cookie")
			}
			return event
		},
	})
	if err != nil {
		log.Fatalf("Sentry init failed: %v", err)
	}

	log.Printf("Sentry initialized — env: %s release: %s", cfg.Environment, cfg.Release)
}

//This middleware automatically captures errors/panics globally, so you do not need to manually capture every crash inside handlers.
func Middleware() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{
		Repanic: true,
	})
}

// CaptureError — for standard context (services, workers)
func CaptureError(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		sentry.CaptureException(err)
	}
}

//This function is used to manually send an error to Sentry.
// CaptureGinError — for Gin handlers use this in handlers
func CaptureGinError(ctx *gin.Context, err error) {
	if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		sentry.CaptureException(err)
	}
}

// CaptureMessage — for standard context
func CaptureMessage(ctx context.Context, msg string) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureMessage(msg)
	} else {
		sentry.CaptureMessage(msg)
	}
}

// CaptureGinMessage — for Gin handlers
func CaptureGinMessage(ctx *gin.Context, msg string) {
	if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
		hub.CaptureMessage(msg)
	} else {
		sentry.CaptureMessage(msg)
	}
}

//This function attaches current logged-in user information to Sentry for the current request in Gin.
func SetUser(ctx *gin.Context, userID, email string) {
	if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetUser(sentry.User{
			ID:    userID,
			Email: email,
		})
	}
}

//This function is used in a background worker system to send worker/job errors to Sentry with extra metadata.
func CaptureWorkerError(err error, taskType, taskID string) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("task_type", taskType)
		scope.SetTag("task_id", taskID)
		scope.SetLevel(sentry.LevelError)
		scope.SetContext("worker", map[string]any{
			"name": "asynq",
			"task": taskType,
			"id":   taskID,
		})
		sentry.CaptureException(err)
	})
}

//If panic happens, report it to Sentry before app crashes.
// fixed — removed panic(r)
func RecoverWithSentry() {
	if r := recover(); r != nil {
		sentry.CurrentHub().Recover(r)
		sentry.Flush(2 * time.Second)
		panic(r)
	}
}

//Wrap all HTTP requests with Sentry monitoring so panics/errors are captured automatically.
func HTTPMiddleware(next http.Handler) http.Handler {
	return sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	}).Handle(next)
}

//Wait up to 3 seconds so Sentry can finish delivering already-sent errors/messages.
func Flush() {
	sentry.Flush(3 * time.Second)
	log.Println("Sentry flushed")
}
