package shared

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthenticatedContext struct {
	*gin.Context
	UserId int64
}

type AuthenticatedHandlerFunc func(*AuthenticatedContext)

func RequireAuth(handler AuthenticatedHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		userId, ok := session.Get("userId").(int64)
		if !ok {
			c.JSON(http.StatusUnauthorized, authenticationRequiredErrorResponse)
			return
		}

		handler(&AuthenticatedContext{
			Context: c,
			UserId:  userId,
		})
	}
}

var authenticationRequiredErrorResponse = ErrorResponse{
	ErrorCode: "authentication_required",
	Message:   "This endpoint requires authentication. Please sign-in.",
}
