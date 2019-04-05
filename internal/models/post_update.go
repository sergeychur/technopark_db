package models

type PostUpdate struct {

	// Собственно сообщение форума.
	Message string `json:"message,omitempty"`
}
