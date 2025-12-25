package service

import (
	"encoding/json"

	"avatar-face-swap-go/internal/database"
)

// LogEntry represents a log record
type LogEntry struct {
	ID        int    `json:"id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Module    string `json:"module"`
	Action    string `json:"action"`
	UserID    string `json:"user_id,omitempty"`
	EventID   string `json:"event_id,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Details   any    `json:"details,omitempty"`
}

func LogActivity(level, module, action, userID, eventID, ipAddress string, details map[string]any) error {
	var detailsJSON *string
	if details != nil {
		data, err := json.Marshal(details)
		if err == nil {
			s := string(data)
			detailsJSON = &s
		}
	}

	query := `INSERT INTO system_log (level, module, action, user_id, event_id, ip_address, details) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := database.DB.Exec(query, level, module, action, userID, eventID, ipAddress, detailsJSON)
	return err
}

func GetLogs(page, perPage int, level, module, startDate, endDate string) ([]LogEntry, int, error) {
	// Build query with filters
	query := `SELECT id, timestamp, level, module, action, user_id, event_id, ip_address, details 
              FROM system_log WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM system_log WHERE 1=1`
	var args []any

	if level != "" {
		query += " AND level = ?"
		countQuery += " AND level = ?"
		args = append(args, level)
	}
	if module != "" {
		query += " AND module = ?"
		countQuery += " AND module = ?"
		args = append(args, module)
	}
	if startDate != "" {
		query += " AND timestamp >= ?"
		countQuery += " AND timestamp >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND timestamp <= ?"
		countQuery += " AND timestamp <= ?"
		args = append(args, endDate)
	}

	var total int
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// pagination
	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, perPage, (page-1)*perPage)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var log LogEntry
		var userID, eventID, ipAddress, details *string

		err := rows.Scan(&log.ID, &log.Timestamp, &log.Level, &log.Module, &log.Action,
			&userID, &eventID, &ipAddress, &details)
		if err != nil {
			return nil, 0, err
		}

		if userID != nil {
			log.UserID = *userID
		}
		if eventID != nil {
			log.EventID = *eventID
		}
		if ipAddress != nil {
			log.IPAddress = *ipAddress
		}
		if details != nil {
			json.Unmarshal([]byte(*details), &log.Details)
		}

		logs = append(logs, log)
	}

	return logs, total, rows.Err()
}
