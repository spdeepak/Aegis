package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/spdeepak/aegis/server/internal/error"
)

func ErrorMiddleware(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			slog.ErrorContext(c, "Panic occurred", "error", err, "path", c.Request.URL.Path)
			// Respond with an error to the client
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				httperror.HttpError{
					Description: "Internal error",
					ErrorCode:   "500",
					Metadata:    fmt.Sprintf("%v", err),
					StatusCode:  http.StatusInternalServerError,
				},
			)
		}
	}()
	c.Next()
}
