package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handlers) registerIndex(e *gin.Engine) {
	e.GET("/", h.getIndex())
}

func (h *Handlers) getIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(301, "/credentials")
	}
}
