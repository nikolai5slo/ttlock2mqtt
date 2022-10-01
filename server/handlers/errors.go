package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) renderInternalError(c *gin.Context, err error) {
	c.HTML(http.StatusInternalServerError, "error.html", gin.H{
		"message": err.Error(),
	})
}
