package handler

import (
	"strconv"

	"avatar-face-swap-go/internal/model"
	"avatar-face-swap-go/internal/repository"
	"avatar-face-swap-go/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}
	event, err := repository.GetEventById(id)
	if err != nil {
		response.Error(c, 500, "Database error")
		return
	}

	if event == nil {
		response.Error(c, 404, "Event not found")
		return
	}

	event.Token = ""
	response.Success(c, event)
}

func ListEvents(c *gin.Context) {
	events, err := repository.GetAllEvents()
	if err != nil {
		response.Error(c, 500, "Database error")
		return
	}

	response.Success(c, gin.H{"events": events})
}

func CreateEvent(c *gin.Context) {
	var req model.CreateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid request: "+err.Error())
		return
	}

	creator, _ := c.Get("user_email")
	creatorStr, _ := creator.(string)

	id, err := repository.CreateEvent(&req, creatorStr)
	if err != nil {
		response.Error(c, 500, "Failed to create event")
		return
	}

	response.Created(c, gin.H{
		"message":  "Event created",
		"event_id": id,
	})
}
