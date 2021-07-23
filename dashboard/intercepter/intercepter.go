package intercepter

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

func MiddleWareCors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool { // 所有
			return true
		},
		MaxAge: 12 * time.Hour,
	})
}

func MiddleWareLog() gin.HandlerFunc {
	if os.Getenv("ENV") == "local" { // text formatter
		return gin.Logger()
	}
	// json formatter
	return gin.LoggerWithFormatter(jsonFormatLogger)
}

func jsonFormatLogger(params gin.LogFormatterParams) string {
	res, _ := jsoniter.MarshalToString(params)
	return res
}
