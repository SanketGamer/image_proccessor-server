package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

//basically bar bar redis pe req karu ek user bohat req banega thats why take a limit.if cnt > limit reject
func RateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc{
	return func(ctx *gin.Context){
		userID,exists:=ctx.Get("userID")
		if !exists{
			ctx.Next()
			return 
		}

        // Build a unique Redis key per user, ex: "ratelimit:user:abc123"
		key := fmt.Sprintf("ratelimit:user:%v", userID)	
		// incr func increments the counter for this key
        // If key doesn't exist yet, Redis creates it starting at 1
		count,err:=rdb.Incr(context.Background(),key).Result()
		if err!=nil{
			ctx.Next()  //if redis fail do not block the user
			return
		}
		//on first req, set the expiry window means what time
		if count == 1 {
			rdb.Expire(context.Background(),key,window)
		}
		//if over limit-> reject
		if count > int64(limit){
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": fmt.Sprintf("Rate limit exceeded. Max %d requests per %s", limit, window),
            })
			return
		}
		ctx.Next()
	}
}