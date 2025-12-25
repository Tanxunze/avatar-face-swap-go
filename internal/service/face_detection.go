package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"

	"avatar-face-swap-go/internal/config"
	"avatar-face-swap-go/internal/storage"

	"golang.org/x/image/draw"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	iai "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/iai/v20200303"
)

type FaceDetectionResult struct {
	Filename    string         `json:"filename"`
	Coordinates FaceCoordinate `json:"coordinates"`
	Confidence  float64        `json:"confidence,omitempty"`
}

type FaceCoordinate struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

type DetectFacesResult struct {
	ImageInfo ImageInfo             `json:"image_info"`
	Faces     []FaceDetectionResult `json:"faces"`
}

type ImageInfo struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Filename string `json:"filename"`
}

func DetectFaces(imagePath string) (*DetectFacesResult, error) {
	cfg := config.Load()

	log.Printf("[FaceDetection] Starting detection for: %s", imagePath)

	if cfg.TencentSecretID == "" || cfg.TencentSecretKey == "" {
		return nil, fmt.Errorf("Tencent Cloud credentials not configured")
	}

	log.Printf("[FaceDetection] Credentials loaded, SecretID prefix: %s...", cfg.TencentSecretID[:10])

	// Read and encode image
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	log.Printf("[FaceDetection] Image size: %d bytes", len(imageData))

	// Decode image to check dimensions
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	origWidth, origHeight := bounds.Dx(), bounds.Dy()
	log.Printf("[FaceDetection] Original dimensions: %dx%d", origWidth, origHeight)

	// Tencent API limits: JPG max 4000px, others max 2000px on long edge
	maxEdge := 4000
	longEdge := max(origWidth, origHeight)

	var scale float64 = 1.0
	if longEdge > maxEdge {
		scale = float64(maxEdge) / float64(longEdge)
		newWidth := int(float64(origWidth) * scale)
		newHeight := int(float64(origHeight) * scale)
		log.Printf("[FaceDetection] Resizing to %dx%d (scale: %.2f)", newWidth, newHeight, scale)

		img = resizeImage(img, newWidth, newHeight)

		// Re-encode to JPEG
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, fmt.Errorf("failed to encode resized image: %w", err)
		}
		imageData = buf.Bytes()
		log.Printf("[FaceDetection] Resized image size: %d bytes", len(imageData))
	}

	// Check if image is too large (5MB limit for base64)
	if len(imageData) > 5*1024*1024 {
		return nil, fmt.Errorf("image too large: %d bytes (max 5MB)", len(imageData))
	}

	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	// Create Tencent Cloud client
	credential := common.NewCredential(cfg.TencentSecretID, cfg.TencentSecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "iai.tencentcloudapi.com"

	client, err := iai.NewClient(credential, cfg.TencentRegion, cpf)
	if err != nil {
		return nil, fmt.Errorf("failed to create Tencent client: %w", err)
	}

	// Build request
	request := iai.NewDetectFaceRequest()
	maxFaceNum := uint64(120)
	minFaceSize := uint64(34)
	faceModelVersion := "3.0"

	request.Image = &imageBase64
	request.MaxFaceNum = &maxFaceNum
	request.MinFaceSize = &minFaceSize
	request.FaceModelVersion = &faceModelVersion

	// Call API
	log.Printf("[FaceDetection] Calling Tencent API...")
	response, err := client.DetectFace(request)
	if err != nil {
		log.Printf("[FaceDetection] API call failed: %v", err)
		return nil, fmt.Errorf("Tencent API error: %w", err)
	}

	log.Printf("[FaceDetection] API call successful, found %d faces", len(response.Response.FaceInfos))

	// Parse result - use ORIGINAL dimensions for metadata
	result := &DetectFacesResult{
		ImageInfo: ImageInfo{
			Width:    origWidth,
			Height:   origHeight,
			Filename: "original.jpg",
		},
		Faces: make([]FaceDetectionResult, 0),
	}

	apiWidth := int(*response.Response.ImageWidth)
	apiHeight := int(*response.Response.ImageHeight)

	for i, face := range response.Response.FaceInfos {
		x := int(*face.X)
		y := int(*face.Y)
		w := int(*face.Width)
		h := int(*face.Height)

		// Scale coordinates back to original image size
		if scale < 1.0 {
			x = int(float64(x) / scale)
			y = int(float64(y) / scale)
			w = int(float64(w) / scale)
			h = int(float64(h) / scale)
		}

		// Add padding (scaled)
		padding := 10
		if scale < 1.0 {
			padding = int(float64(padding) / scale)
		}
		x1 := max(0, x-padding)
		y1 := max(0, y-padding)
		x2 := min(origWidth, x+w+padding)
		y2 := min(origHeight, y+h+padding)

		log.Printf("[FaceDetection] Face %d: API coords (%d,%d,%d,%d) -> Original coords (%d,%d,%d,%d)",
			i+1, int(*face.X), int(*face.Y), int(*face.Width), int(*face.Height), x1, y1, x2, y2)

		result.Faces = append(result.Faces, FaceDetectionResult{
			Filename: fmt.Sprintf("face_%d.jpg", i+1),
			Coordinates: FaceCoordinate{
				X1: x1,
				Y1: y1,
				X2: x2,
				Y2: y2,
			},
			Confidence: 1.0,
		})
	}

	_ = apiWidth // suppress unused warning
	_ = apiHeight

	return result, nil
}

func ProcessEventImage(eventID int, imagePath string) error {
	log.Printf("[ProcessEventImage] Starting for event %d, path: %s", eventID, imagePath)

	// Detect faces
	result, err := DetectFaces(imagePath)
	if err != nil {
		log.Printf("[ProcessEventImage] DetectFaces failed: %v", err)
		return err
	}

	log.Printf("[ProcessEventImage] Detected %d faces", len(result.Faces))

	// Ensure directories exist
	if err := storage.EnsureEventDirs(eventID); err != nil {
		return err
	}

	// Open original image for cropping
	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Crop and save each face
	for _, face := range result.Faces {
		croppedImg := cropImageRect(img, face.Coordinates.X1, face.Coordinates.Y1, face.Coordinates.X2, face.Coordinates.Y2)

		facePath := storage.GetFacePath(eventID, face.Filename)
		faceFile, err := os.Create(facePath)
		if err != nil {
			return err
		}

		if err := jpeg.Encode(faceFile, croppedImg, &jpeg.Options{Quality: 90}); err != nil {
			faceFile.Close()
			return err
		}
		faceFile.Close()
	}

	// Copy original to storage
	originalDest := storage.GetOriginalPath(eventID)
	if imagePath != originalDest {
		srcData, err := os.ReadFile(imagePath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(originalDest, srcData, 0644); err != nil {
			return err
		}
	}

	// Save metadata
	return saveMetadata(eventID, result)
}

func resizeImage(img image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

func cropImageRect(img image.Image, x1, y1, x2, y2 int) image.Image {
	if subImager, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		return subImager.SubImage(image.Rect(x1, y1, x2, y2))
	}

	// Fallback
	cropped := image.NewRGBA(image.Rect(0, 0, x2-x1, y2-y1))
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			cropped.Set(x-x1, y-y1, img.At(x, y))
		}
	}
	return cropped
}

func saveMetadata(eventID int, result *DetectFacesResult) error {
	metadataPath := storage.GetMetadataPath(eventID)

	// Convert to map for JSON
	metadata := map[string]any{
		"image_info": result.ImageInfo,
		"faces":      result.Faces,
	}

	data, err := json.MarshalIndent(metadata, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}
