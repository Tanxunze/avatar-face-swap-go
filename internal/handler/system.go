package handler

import (
	"avatar-face-swap-go/internal/config"
	"avatar-face-swap-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetDetectorStatus
// GET /api/system/detector-status
// Reports the face-detection backend status. The Go implementation uses the
// Tencent Cloud IAI service (no local model / GPU), so this mainly reports
// whether the cloud credentials are configured. Response shape mirrors the
// Python backend so the admin debug panel works against either backend.
func GetDetectorStatus(c *gin.Context) {
	cfg := config.Load()

	configured := cfg.TencentSecretID != "" && cfg.TencentSecretKey != ""

	loadedDetectors := []string{}
	if configured {
		loadedDetectors = append(loadedDetectors, "tencent-cloud-iai")
	}

	detectorStatus := gin.H{
		"backend":            "tencent-cloud-iai",
		"loaded_detectors":   loadedDetectors,
		"gpu_available":      false, // cloud detection, not applicable
		"available":          configured,
		"region":             cfg.TencentRegion,
		"face_model_version": "3.0",
	}

	message := "检测器状态获取成功（腾讯云 IAI）"
	if !configured {
		message = "未配置腾讯云凭证，云端人脸检测不可用"
	}

	response.Success(c, gin.H{
		"success":         true,
		"detector_status": detectorStatus,
		"message":         message,
	})
}
