package handler

import (
	"strconv"

	"avatar-face-swap-go/internal/service"
	"avatar-face-swap-go/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	level := c.Query("level")
	module := c.Query("module")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	logs, total, err := service.GetLogs(page, perPage, level, module, startDate, endDate)
	if err != nil {
		response.Error(c, 500, "Failed to get logs")
		return
	}

	response.Success(c, gin.H{
		"logs":     logs,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}
