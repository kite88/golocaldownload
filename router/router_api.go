package router

import (
	"github.com/gin-gonic/gin"

	"golocaldownload/handle"
)

// apiR api 路由。
func apiR(r *gin.Engine, h *handle.Handler) {
	api := r.Group("/api")
	{
		api.GET("/list", h.List)
		api.GET("/download", h.Download)
		api.POST("/search", h.Search)
	}
}
