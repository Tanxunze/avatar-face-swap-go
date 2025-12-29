package handler

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // PNG decoder
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"avatar-face-swap-go/internal/service"
	"avatar-face-swap-go/internal/storage"
	"avatar-face-swap-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// GET /api/events/:id/picture
// Returns the main event picture
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

// GET /api/events/:id/faces/metadata
// Returns metadata for all detected faces
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

// PUT /api/events/:id/picture
// Uploads or replaces the main event picture (triggers face detection)
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

	// Async face detection
	go func() {
		if err := service.ProcessEventImage(eventID, destPath); err != nil {
			fmt.Printf("Face detection failed for event %d: %v\n", eventID, err)
			service.LogActivity("ERROR", "图片处理", "人脸识别失败", "", strconv.Itoa(eventID), "", map[string]any{
				"error": err.Error(),
			})
		} else {
			service.LogActivity("INFO", "图片处理", "人脸识别完成", "", strconv.Itoa(eventID), "", nil)
		}
	}()

	c.JSON(202, gin.H{
		"success": true,
		"data": gin.H{
			"message":  "Image uploaded, processing faces",
			"event_id": eventID,
		},
	})
}

// POST /api/events/:id/faces/:face/avatar
// Uploads a custom avatar image for a specific face
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

// GET /api/events/:id/qq-profiles/:qq
// Returns QQ nickname for a given QQ number
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

// POST /api/events/:id/faces/:face/qq-avatar
// Downloads QQ avatar and saves it for a specific face
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

	response.Success(c, gin.H{"message": "头像上传完成"})
}

// GET /api/events/:id/faces/:filename/qq-profile
// Returns QQ profile info associated with a face
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

// GET /api/events/:id/avatars/:filename
// Returns an uploaded avatar image
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

// GET /api/events/:id/picture/metadata
// Returns metadata about the event picture (dimensions, etc.)
func GetEventPicInfo(c *gin.Context) {
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

	imageInfo, ok := metadata["image_info"]
	if !ok {
		response.Error(c, 404, "Image info not found")
		return
	}

	response.Success(c, gin.H{
		"pic_info": imageInfo,
		"event_id": eventID,
	})
}

// POST /api/events/:id/faces
// Manually adds a face by specifying coordinates to crop from the main picture
func AddManualFace(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "Invalid event ID")
		return
	}

	var req struct {
		X1     int    `json:"x1" binding:"required"`
		Y1     int    `json:"y1" binding:"required"`
		X2     int    `json:"x2" binding:"required"`
		Y2     int    `json:"y2" binding:"required"`
		FaceID string `json:"face_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid request: "+err.Error())
		return
	}

	// Generate face_id if not provided
	if req.FaceID == "" {
		req.FaceID = fmt.Sprintf("manual_%d", time.Now().UnixMilli())
	}

	// Read original image
	originalPath := storage.GetOriginalPath(eventID)
	originalFile, err := os.Open(originalPath)
	if err != nil {
		response.Error(c, 404, "Original image not found")
		return
	}
	defer originalFile.Close()

	img, _, err := image.Decode(originalFile)
	if err != nil {
		response.Error(c, 500, "Failed to decode image")
		return
	}

	// Get image bounds
	bounds := img.Bounds()
	imgW, imgH := bounds.Max.X, bounds.Max.Y

	// Clamp coordinates
	x1 := clamp(req.X1, 0, imgW)
	y1 := clamp(req.Y1, 0, imgH)
	x2 := clamp(req.X2, x1, imgW)
	y2 := clamp(req.Y2, y1, imgH)

	// Crop image
	croppedImg := cropImage(img, x1, y1, x2, y2)

	// Ensure directories exist
	if err := storage.EnsureEventDirs(eventID); err != nil {
		response.Error(c, 500, "Failed to create directories")
		return
	}

	// Save cropped face
	faceFilename := req.FaceID + ".jpg"
	facePath := storage.GetFacePath(eventID, faceFilename)

	faceFile, err := os.Create(facePath)
	if err != nil {
		response.Error(c, 500, "Failed to create face file")
		return
	}
	defer faceFile.Close()

	if err := jpeg.Encode(faceFile, croppedImg, &jpeg.Options{Quality: 90}); err != nil {
		response.Error(c, 500, "Failed to save face image")
		return
	}

	// Update metadata.json
	newFace := map[string]any{
		"filename": faceFilename,
		"coordinates": map[string]int{
			"x1": x1,
			"y1": y1,
			"x2": x2,
			"y2": y2,
		},
		"confidence": 1.0,
		"manual":     true,
		"face_id":    req.FaceID,
	}

	if err := addFaceToMetadata(eventID, imgW, imgH, newFace); err != nil {
		fmt.Printf("Warning: failed to update metadata: %v\n", err)
	}

	userEmail, _ := c.Get("user_email")
	userEmailStr, _ := userEmail.(string)

	service.LogActivity("INFO", "图片处理", "手动添加人脸", userEmailStr, strconv.Itoa(eventID), c.ClientIP(), map[string]any{
		"face_id":     req.FaceID,
		"coordinates": newFace["coordinates"],
	})

	response.Created(c, gin.H{
		"message":   "Face added",
		"face_info": newFace,
	})
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func cropImage(img image.Image, x1, y1, x2, y2 int) image.Image {
	rect := image.Rect(x1, y1, x2, y2)

	// SubImage if supported
	if subImager, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		return subImager.SubImage(rect)
	}

	// Fallback: manual crop
	cropped := image.NewRGBA(image.Rect(0, 0, x2-x1, y2-y1))
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			cropped.Set(x-x1, y-y1, img.At(x, y))
		}
	}
	return cropped
}

func addFaceToMetadata(eventID, imgW, imgH int, newFace map[string]any) error {
	metadataPath := storage.GetMetadataPath(eventID)

	var metadata map[string]any

	data, err := os.ReadFile(metadataPath)
	if os.IsNotExist(err) {
		// Create new metadata
		metadata = map[string]any{
			"image_info": map[string]any{
				"width":    imgW,
				"height":   imgH,
				"filename": "original.jpg",
			},
			"faces": []any{},
		}
	} else if err != nil {
		return err
	} else {
		if err := json.Unmarshal(data, &metadata); err != nil {
			return err
		}
	}

	// Append new face
	faces, ok := metadata["faces"].([]any)
	if !ok {
		faces = []any{}
	}
	metadata["faces"] = append(faces, newFace)

	// Write back
	updated, err := json.MarshalIndent(metadata, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, updated, 0644)
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
