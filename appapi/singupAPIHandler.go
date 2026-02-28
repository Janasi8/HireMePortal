package signupapiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitializeAPI(router *gin.Engine) {

	// ✅ REAL LOGIN (DB-based)
	router.POST("/login", LoginGinHandler)

	// ✅ SIGNUP (keep as-is for now)
	router.POST("/signup", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "signup success",
		})
	})
}
