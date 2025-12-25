package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"avatar-face-swap-go/internal/config"
)

func GetEventDir(eventID int) string {
	cfg := config.Load()
	return filepath.Join(cfg.StorageDir, "events", fmt.Sprintf("%d", eventID))
}

func GetOriginalPath(eventID int) string {
	return filepath.Join(GetEventDir(eventID), "original.jpg")
}

func GetMetadataPath(eventID int) string {
	return filepath.Join(GetEventDir(eventID), "metadata.json")
}

func GetFacesDir(eventID int) string {
	return filepath.Join(GetEventDir(eventID), "faces")
}

func GetFacePath(eventID int, filename string) string {
	return filepath.Join(GetFacesDir(eventID), filename)
}

func GetAvatarsDir(eventID int) string {
	return filepath.Join(GetEventDir(eventID), "avatars")
}

func GetAvatarPath(eventID int, filename string) string {
	return filepath.Join(GetAvatarsDir(eventID), filename)
}

func EnsureEventDirs(eventID int) error {
	dirs := []string{
		GetEventDir(eventID),
		GetFacesDir(eventID),
		GetAvatarsDir(eventID),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
