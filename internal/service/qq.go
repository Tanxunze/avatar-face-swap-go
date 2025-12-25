package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"avatar-face-swap-go/internal/storage"
)

type QQNameResponse struct {
	Code int    `json:"code"`
	Name string `json:"name"`
}

func GetQQNickname(qqNumber string) (string, error) {
	url := fmt.Sprintf("https://api.ilingku.com/int/v1/qqname?qq=%s", qqNumber)

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result QQNameResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Code == 200 && result.Name != "" {
		return result.Name, nil
	}

	return fmt.Sprintf("QQ用户%s", qqNumber), nil
}

func DownloadQQAvatar(eventID int, face, qqNumber string) error {
	url := fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=640", qqNumber)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download avatar: %d", resp.StatusCode)
	}

	if err := storage.EnsureEventDirs(eventID); err != nil {
		return err
	}

	baseName := face[:len(face)-len(".jpg")]
	avatarPath := storage.GetAvatarPath(eventID, baseName+".jpg")

	file, err := os.Create(avatarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	// QQ info JSON
	jsonPath := storage.GetAvatarPath(eventID, baseName+".json")
	info := map[string]string{
		"qq_number": qqNumber,
		"filename":  baseName + ".jpg",
	}
	jsonData, _ := json.Marshal(info)

	return os.WriteFile(jsonPath, jsonData, 0644)
}
