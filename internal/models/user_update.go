package models

// UserUpdate Информация о пользователе.
//
// swagger:model UserUpdate
type UserUpdate struct {

	// Описание пользователя.
	About string `json:"about,omitempty"`

	// Почтовый адрес пользователя (уникальное поле).
	Email strfmt.Email `json:"email,omitempty"`

	// Полное имя пользователя.
	Fullname string `json:"fullname,omitempty"`
}
