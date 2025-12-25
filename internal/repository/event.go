package repository

import (
	"database/sql"
	"strings"

	"avatar-face-swap-go/internal/database"
	"avatar-face-swap-go/internal/model"
)

func GetEventByID(id int) (*model.Event, error) {
	query := `SELECT event_id, description, token, event_date, is_open, creator 
              FROM event WHERE event_id = ?`
	var event model.Event
	var creator sql.NullString

	err := database.DB.QueryRow(query, id).Scan(
		&event.ID,
		&event.Description,
		&event.Token,
		&event.EventDate,
		&event.IsOpen,
		&creator,
	)

	if err == sql.ErrNoRows {
		return nil, nil // not found, no error
	}
	if err != nil {
		return nil, err
	}

	if creator.Valid {
		event.Creator = creator.String
	}

	return &event, nil
}

func GetAllEvents() ([]model.Event, error) {
	query := `SELECT event_id, description, event_date, is_open FROM event`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.Description, &e.EventDate, &e.IsOpen); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

func CreateEvent(req *model.CreateEventRequest, creator string) (int64, error) {
	query := `INSERT INTO event (description, token, event_date, is_open, creator) 
              VALUES (?, ?, ?, ?, ?)`
	isOpen := 0
	if req.IsOpen {
		isOpen = 1
	}

	result, err := database.DB.Exec(query,
		req.Description,
		req.Token,
		req.EventDate,
		isOpen,
		creator,
	)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func GetEventByToken(token string) (*model.Event, error) {
	query := `SELECT event_id, description, token, event_date, is_open, creator 
              FROM event WHERE token = ?`

	var event model.Event
	var creator sql.NullString

	err := database.DB.QueryRow(query, token).Scan(
		&event.ID,
		&event.Description,
		&event.Token,
		&event.EventDate,
		&event.IsOpen,
		&creator,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if creator.Valid {
		event.Creator = creator.String
	}

	return &event, nil
}

func UpdateEvent(id int, req *model.UpdateEventRequest) error {
	var fields []string
	var args []any

	if req.Description != nil {
		fields = append(fields, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Token != nil {
		fields = append(fields, "token = ?")
		args = append(args, *req.Token)
	}
	if req.EventDate != nil {
		fields = append(fields, "event_date = ?")
		args = append(args, *req.EventDate)
	}
	if req.IsOpen != nil {
		isOpen := 0
		if *req.IsOpen {
			isOpen = 1
		}
		fields = append(fields, "is_open = ?")
		args = append(args, isOpen)
	}

	if len(fields) == 0 {
		return nil
	}

	query := "UPDATE event SET " + strings.Join(fields, ", ") + " WHERE event_id = ?"
	args = append(args, id)

	_, err := database.DB.Exec(query, args...)
	return err
}

func DeleteEvent(id int) error {
	_, err := database.DB.Exec("DELETE FROM event WHERE event_id = ?", id)
	return err
}
