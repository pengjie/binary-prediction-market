package response

import (
	"github.com/gin-gonic/gin"
	"github.com/huinong/golang-claude/pkg/common/errors"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func OK(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, err error) {
	appErr, ok := err.(*errors.AppError)
	if ok {
		c.JSON(200, Response{
			Success: false,
			Message: appErr.Message,
		})
	} else {
		c.JSON(200, Response{
			Success: false,
			Message: err.Error(),
		})
	}
}