package router

import (
	"github.com/gin-gonic/gin"
	"golocaldownload/handle"
)

// ApiR api路由
func apiR(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/list", handle.Ins.List)
		api.GET("/download", handle.Ins.Download)
		api.POST("/search", handle.Ins.Search)
	}
}
