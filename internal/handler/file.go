package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"avatar-face-swap-go/internal/service"
	"avatar-face-swap-go/internal/storage"
	"avatar-face-swap-go/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetEventPic(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	imagePath := storage.GetOriginalPath(eventID)

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		response.Error(c, 404, "Image not found")
		return
	}

	c.File(imagePath)
}

func GetEventFaces(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	facesDir := storage.GetFacesDir(eventID)

	if _, err := os.Stat(facesDir); os.IsNotExist(err) {
		response.Error(c, 404, "Faces directory not found")
		return
	}

	entries, err := os.ReadDir(facesDir)
	if err != nil {
		response.Error(c, 500, "Failed to read faces directory")
		return
	}

	var faces []string
	for _, entry := range entries {
		if !entry.IsDir() {
			faces = append(faces, entry.Name())
		}
	}

	response.Success(c, gin.H{
		"faces":    faces,
		"event_id": eventID,
	})
}

func GetFaceImage(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	filename := c.Param("filename")

	// prevent directory traversal
	if filepath.Base(filename) != filename {
		response.Error(c, 400, "Invalid filename")
		return
	}

	facePath := storage.GetFacePath(eventID, filename)

	if _, err := os.Stat(facePath); os.IsNotExist(err) {
		response.Error(c, 404, "Face image not found")
		return
	}

	c.File(facePath)
}

func GetEventMetadata(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	metadataPath := storage.GetMetadataPath(eventID)

	data, err := os.ReadFile(metadataPath)
	if os.IsNotExist(err) {
		response.Error(c, 404, "Metadata not found")
		return
	}
	if err != nil {
		response.Error(c, 500, "Failed to read metadata")
		return
	}

	var metadata map[string]any
	if err := json.Unmarshal(data, &metadata); err != nil {
		response.Error(c, 500, "Invalid metadata format")
		return
	}

	response.Success(c, metadata)
}

func UploadEventPic(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, "No file uploaded")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		response.Error(c, 400, "Only jpg, jpeg, png allowed")
		return
	}

	if err := storage.EnsureEventDirs(eventID); err != nil {
		response.Error(c, 500, "Failed to create directories")
		return
	}

	destPath := storage.GetOriginalPath(eventID)
	if err := c.SaveUploadedFile(file, destPath); err != nil {
		response.Error(c, 500, "Failed to save file")
		return
	}

	userEmail, _ := c.Get("user_email")
	userEmailStr, _ := userEmail.(string)

	service.LogActivity("INFO", "活动管理", "上传活动图片", userEmailStr, strconv.Itoa(eventID), c.ClientIP(), map[string]any{
		"filename": file.Filename,
	})

	response.Success(c, gin.H{
		"message":  "Image uploaded",
		"event_id": eventID,
	})
}

func UploadAvatar(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	face := c.Param("face")
	if filepath.Base(face) != face {
		response.Error(c, 400, "Invalid face parameter")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, "No file uploaded")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		response.Error(c, 400, "Only jpg, jpeg, png allowed")
		return
	}

	if err := storage.EnsureEventDirs(eventID); err != nil {
		response.Error(c, 500, "Failed to create directories")
		return
	}

	baseName := face[:len(face)-len(filepath.Ext(face))]
	destPath := storage.GetAvatarPath(eventID, baseName+ext)

	if err := c.SaveUploadedFile(file, destPath); err != nil {
		response.Error(c, 500, "Failed to save file")
		return
	}

	response.Success(c, gin.H{
		"message":  "Avatar uploaded",
		"filename": baseName + ext,
	})
}

func GetQQNickname(c *gin.Context) {
	qqNumber := c.Param("qq")

	nickname, err := service.GetQQNickname(qqNumber)
	if err != nil {
		response.Success(c, gin.H{
			"nickname":  fmt.Sprintf("QQ用户%s", qqNumber),
			"qq_number": qqNumber,
			"success":   false,
			"error":     err.Error(),
		})
		return
	}

	response.Success(c, gin.H{
		"nickname":  nickname,
		"qq_number": qqNumber,
		"success":   true,
	})
}

func UploadQQAvatar(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	face := c.Param("face")

	var req struct {
		QQNumber string `json:"qqNumber" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Missing qqNumber")
		return
	}

	service.LogActivity("INFO", "图片处理", "上传QQ头像", "", strconv.Itoa(eventID), c.ClientIP(), map[string]any{
		"face":      face,
		"qq_number": req.QQNumber,
	})

	go func() {
		if err := service.DownloadQQAvatar(eventID, face, req.QQNumber); err != nil {
			fmt.Printf("Failed to download QQ avatar: %v\n", err)
		}
	}()

	response.Success(c, gin.H{"message": "Avatar upload started"})
}

func GetFaceQQInfo(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	filename := c.Param("filename")
	if filepath.Base(filename) != filename {
		response.Error(c, 400, "Invalid filename")
		return
	}

	baseName := filename[:len(filename)-len(filepath.Ext(filename))]
	jsonPath := storage.GetAvatarPath(eventID, baseName+".json")

	data, err := os.ReadFile(jsonPath)
	if os.IsNotExist(err) {
		response.Success(c, gin.H{
			"qq_number": nil,
			"filename":  filename,
		})
		return
	}
	if err != nil {
		response.Error(c, 500, "Failed to read QQ info")
		return
	}

	var info map[string]any
	if err := json.Unmarshal(data, &info); err != nil {
		response.Error(c, 500, "Invalid QQ info format")
		return
	}

	info["filename"] = filename
	response.Success(c, info)
}

func GetUploadedAvatar(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	filename := c.Param("filename")
	if filepath.Base(filename) != filename {
		response.Error(c, 400, "Invalid filename")
		return
	}

	avatarPath := storage.GetAvatarPath(eventID, filename)

	if _, err := os.Stat(avatarPath); os.IsNotExist(err) {
		response.Error(c, 404, "Avatar not found")
		return
	}

	c.File(avatarPath)
}

func DeleteFace(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	filename := c.Param("filename")
	if filepath.Base(filename) != filename {
		response.Error(c, 400, "Invalid filename")
		return
	}

	baseName := filename[:len(filename)-len(filepath.Ext(filename))]

	// Delete face image
	facePath := storage.GetFacePath(eventID, filename)
	if err := os.Remove(facePath); err != nil && !os.IsNotExist(err) {
		response.Error(c, 500, "Failed to delete face image")
		return
	}

	// Delete associated avatar (try all extensions)
	for _, ext := range []string{".jpg", ".jpeg", ".png"} {
		avatarPath := storage.GetAvatarPath(eventID, baseName+ext)
		os.Remove(avatarPath)
	}

	// Delete associated JSON
	jsonPath := storage.GetAvatarPath(eventID, baseName+".json")
	os.Remove(jsonPath)

	// Update metadata.json
	if err := removeFaceFromMetadata(eventID, filename); err != nil {
		fmt.Printf("Warning: failed to update metadata: %v\n", err)
	}

	userEmail, _ := c.Get("user_email")
	userEmailStr, _ := userEmail.(string)

	service.LogActivity("WARNING", "活动管理", "删除误识别人脸", userEmailStr, strconv.Itoa(eventID), c.ClientIP(), map[string]any{
		"deleted_face": filename,
	})

	response.Success(c, gin.H{"message": "Face deleted"})
}

func removeFaceFromMetadata(eventID int, filename string) error {
	metadataPath := storage.GetMetadataPath(eventID)

	data, err := os.ReadFile(metadataPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var metadata map[string]any
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}

	if faces, ok := metadata["faces"].([]any); ok {
		var filtered []any
		for _, f := range faces {
			if face, ok := f.(map[string]any); ok {
				if face["filename"] != filename {
					filtered = append(filtered, face)
				}
			}
		}
		metadata["faces"] = filtered
	}

	updated, err := json.MarshalIndent(metadata, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, updated, 0644)
}
