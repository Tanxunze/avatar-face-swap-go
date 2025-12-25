package repository

import (
	"database/sql"

	"avatar-face-swap-go/internal/database"
	"avatar-face-swap-go/internal/model"
)

func GetEventById(id int) (*model.Event, error) {
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
