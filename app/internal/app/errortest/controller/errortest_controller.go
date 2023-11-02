package controller

import (
	"app/internal/app/errortest/service"

	"example.com/appbase/pkg/logging"
	"github.com/gin-gonic/gin"
)

type ErrorTestController interface {
	Execute(ctx *gin.Context) (interface{}, error)
}

func New(log logging.Logger, service service.ErrorTestService) ErrorTestController {
	return &errorTestControllerImpl{log: log, service: service}
}

type errorTestControllerImpl struct {
	log     logging.Logger
	service service.ErrorTestService
}

// Execute implements ErrorTestController.
func (c *errorTestControllerImpl) Execute(ctx *gin.Context) (interface{}, error) {
	errorType := ctx.Param("errortype")
	return nil, c.service.Execute(errorType)
}
