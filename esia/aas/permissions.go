package aas

import (
	"encoding/base64"
	"encoding/json"
)

// Permissions - список запрашиваемых прав доступа.
//
// Подробнее см "Методические рекомендации по интеграции с REST API Цифрового профиля",
// раздел "Структура JSON-объекта параметра «permissions»".
type Permissions []Permission

// Permission - разрешение на доступ к ресурсу.
//
// Подробнее см "Методические рекомендации по интеграции с REST API Цифрового профиля",
// раздел "Структура JSON-объекта параметра «permissions»".
type Permission struct {
	ResponsibleObject string              `json:"responsibleObject,omitempty"` // Ответственный объект (название организации)
	Sysname           string              `json:"sysname"`                     // Мнемоника типа согласия
	Expire            int                 `json:"expire,omitempty"`            // Срок, на который будет выдано согласие после утверждения (в минутах)
	Actions           []PermissionAction  `json:"actions"`                     // Перечень мнемоник действий
	Purposes          []PermissionPurpose `json:"purposes"`                    // Перечень мнемоник целей согласия
	Scopes            []PermissionScope   `json:"scopes"`                      // Перечень мнемоник областей доступа
}

// PermissionPurpose - мнемоника цели согласия объекта [Permission].
type PermissionPurpose struct {
	Sysname string `json:"sysname"`
}

// PermissionScope - мнемоника области доступа объекта [Permission].
type PermissionScope struct {
	Sysname string `json:"sysname"`
}

// PermissionAction - мнемоника действия объекта [Permission].
type PermissionAction struct {
	Sysname string `json:"sysname"`
}

// Base64String - кодирует список запрашиваемых разрешений в строку base64
func (p Permissions) Base64String() string {
	j, _ := json.Marshal(p)
	return base64.RawURLEncoding.EncodeToString(j)
}
