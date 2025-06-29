package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ViewR view路由
func viewR(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", func() {})
	})
}
