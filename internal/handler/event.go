package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"avatar-face-swap-go/internal/model"
	"avatar-face-swap-go/internal/repository"
	"avatar-face-swap-go/internal/service"
	"avatar-face-swap-go/internal/storage"
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
	event, err := repository.GetEventByID(id)
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

	service.LogActivity("INFO", "活动管理", "创建活动", creatorStr, strconv.FormatInt(id, 10), c.ClientIP(), map[string]any{
		"description": req.Description,
	})

	response.Created(c, gin.H{
		"message":  "Event created",
		"event_id": id,
	})
}

func UpdateEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	var req model.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid request: "+err.Error())
		return
	}

	event, err := repository.GetEventByID(id)
	if err != nil {
		response.Error(c, 500, "Database error")
		return
	}
	if event == nil {
		response.Error(c, 404, "Event not found")
		return
	}

	if err := repository.UpdateEvent(id, &req); err != nil {
		response.Error(c, 500, "Failed to update event")
		return
	}

	userEmail, _ := c.Get("user_email")
	userEmailStr, _ := userEmail.(string)

	service.LogActivity("INFO", "活动管理", "更新活动", userEmailStr, idStr, c.ClientIP(), nil)

	response.Success(c, gin.H{"message": "Event updated"})
}

func DeleteEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	event, err := repository.GetEventByID(id)
	if err != nil {
		response.Error(c, 500, "Database error")
		return
	}
	if event == nil {
		response.Error(c, 404, "Event not found")
		return
	}

	if err := repository.DeleteEvent(id); err != nil {
		response.Error(c, 500, "Failed to delete event")
		return
	}

	userEmail, _ := c.Get("user_email")
	userEmailStr, _ := userEmail.(string)

	service.LogActivity("WARNING", "活动管理", "删除活动", userEmailStr, idStr, c.ClientIP(), nil)

	response.Success(c, gin.H{"message": "Event deleted"})
}

func GetEventToken(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	event, err := repository.GetEventByID(eventID)
	if err != nil {
		response.Error(c, 500, "Database error")
		return
	}

	if event == nil {
		response.Error(c, 404, "Event not found")
		return
	}

	response.Success(c, gin.H{"token": event.Token})
}

// GET /api/events/:id/status
// Returns the face detection processing status for an event
func GetProcessStatus(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	metadataPath := storage.GetMetadataPath(eventID)
	originalPath := storage.GetOriginalPath(eventID)

	if data, err := os.ReadFile(metadataPath); err == nil {
		var metadata map[string]any
		if err := json.Unmarshal(data, &metadata); err == nil {
			facesCount := 0
			if faces, ok := metadata["faces"].([]any); ok {
				facesCount = len(faces)
			}
			response.Success(c, gin.H{
				"status":      "completed",
				"faces_count": facesCount,
				"message":     fmt.Sprintf("Processing completed, %d faces detected", facesCount),
			})
			return
		}
	}

	if _, err := os.Stat(originalPath); err == nil {
		response.Success(c, gin.H{
			"status":  "processing",
			"message": "Processing in progress",
		})
		return
	}

	response.Error(c, 404, "No image uploaded")
}
