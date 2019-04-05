package models

// Vote Информация о голосовании пользователя.
//
// swagger:model Vote
type Vote struct {

	// Идентификатор пользователя.
	// Required: true
	Nickname string `json:"nickname"`

	// Отданный голос.
	// Required: true
	Voice int32 `json:"voice"`
}
