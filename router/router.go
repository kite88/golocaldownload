package router

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"io/fs"
	"net/http"
)

func R(envMode string, viewFS, staticFs fs.FS) *gin.Engine {
	var r = new(gin.Engine)
	if envMode == gin.ReleaseMode {
		r = gin.New()
	} else {
		r = gin.Default()
	}

	r.SetHTMLTemplate(template.Must(template.ParseFS(viewFS, "web/view/*")))
	r.GET("/web/static/*filepath", func(c *gin.Context) {
		http.FileServer(http.FS(staticFs)).ServeHTTP(c.Writer, c.Request)
	})

	viewR(r)
	apiR(r)

	return r
}
