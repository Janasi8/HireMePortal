package signupapiv1

import (
	"github.com/gin-gonic/gin"
)

func LoginGinHandler(c *gin.Context) {
	handleLoginRequests(c.Writer, c.Request)
}
