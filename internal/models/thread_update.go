package models


// ThreadUpdate Сообщение для обновления ветки обсуждения на форуме.
// Пустые параметры остаются без изменений.
//
// swagger:model ThreadUpdate
type ThreadUpdate struct {

	// Описание ветки обсуждения.
	Message string `json:"message,omitempty"`

	// Заголовок ветки обсуждения.
	Title string `json:"title,omitempty"`
}
