package model

type Event struct {
	ID          int    `json:"event_id"`
	Description string `json:"description"`
	Token       string `json:"token,omitempty"`
	EventDate   string `json:"event_date"`
	IsOpen      bool   `json:"is_open"`
	Creator     string `json:"creator,omitempty"`
}

type CreateEventRequest struct {
	Description string `json:"description" binding:"required"`
	Token       string `json:"token" binding:"required"`
	EventDate   string `json:"event_date" binding:"required"`
	IsOpen      bool   `json:"is_open"`
}

type UpdateEventRequest struct {
	Description string `json:"description" binding:"required"`
	Token       string `json:"token" binding:"required"`
	EventDate   string `json:"event_date" binding:"required"`
	IsOpen      bool   `json:"is_open"`
}
