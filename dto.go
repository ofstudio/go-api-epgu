package apipgu

// dtoErrorResponse - ответ API ЕПГУ при ошибке
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.12",
// "Приложение 4. Ошибки, возвращаемые при запросах к API ЕПГУ"
//
// Пример JSON-ответа при ошибке:
//
//	{
//	  "code": "order_access",
//	  "message": "У пользователя нет прав для работы с текущим заявлением"
//	}
type dtoErrorResponse struct {
	Code    string `json:"code"`    // Код ошибки
	Message string `json:"message"` // Сообщение об ошибке
}

// dtoOrderIdResponse - ответ API ЕПГУ с номером созданного заявления.
type dtoOrderIdResponse struct {
	OrderId int `json:"orderId"`
}

// dtoOrderInfoResponse - ответ API ЕПГУ с детальной информацией по отправленному заявлению.
//
// Подробнее см "Спецификация API ЕПГУ версия 1.12",
// раздел "2.4. Получение деталей по заявлению".
//
// Пример для заявления "Доставка пенсии и социальных выплат СФР" (10000000109):
//
//	{
//	  "code": "OK",
//	  "message": null,
//	  "messageId": "2252fb21-92f8-61ee-a6f0-7ed53c117861",
//	  "order": "{...}"
//	}
type dtoOrderInfoResponse struct {
	Code      string `json:"code"`      // Код состояния заявления в соответствии с Приложением 1 Спецификации
	Message   string `json:"message"`   // Текстовое сообщение, описывающее текущее состояние запроса на создание заявления
	MessageId string `json:"messageId"` // [Не документировано, GUID]
	Order     string `json:"order"`     // В случае, если заявление уже создано на портале и отправлено в ведомство, параметр содержит строку в виде экранированного JSON-объекта
}

// dtoDictRequest - запрос к API ЕПГУ на получение справочника.
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.12",
// раздел "3. Получение справочных данных".
type dtoDictRequest struct {
	TreeFiltering      string `json:"treeFiltering"`                // Тип справочника (плоский / иерархический)
	ParentRefItemValue string `json:"parentRefItemValue,omitempty"` // Код родительского элемента
	PageNum            int    `json:"pageNum,omitempty"`            // Номер необходимой страницы
	PageSize           int    `json:"pageSize,omitempty"`           // Количество записей на странице
}

// dtoDictResponse - ответ API ЕПГУ на запрос справочника.
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.12",
// раздел "3. Получение справочных данных".
//
// Пример структуры успешного ответа:
//
//	{
//	  "error": {
//	    "code": 0,
//	    "message": "operation completed"
//	  },
//	  "fieldErrors": [],
//	  "total": 1011,
//	  "items": [...элементы справочника...]
//	}
//
// Пример структуры ответа в случае ошибки:
//
//	{
//	  "error": {
//	    "code": 7,
//	    "message": "Entity not found"
//	  },
//	  "fieldErrors": [],
//	  "total": 0,
//	  "items": []
//	}
type dtoDictResponse struct {
	Error       dtoDictResponseError     `json:"error"`       // Результат выполнения операции
	FieldErrors []map[string]interface{} `json:"fieldErrors"` // Ошибки в полях запроса
	Total       int                      `json:"total"`       // Общее количество найденных элементов
	Items       []DictItem               `json:"items"`       // Найденные элементы справочника с учётом заданных pageNum и pageSize
}

// dtoDictResponseError - структура поля dtoDictResponse.Error.
type dtoDictResponseError struct {
	Code    int    `json:"code"`    // Код результата
	Message string `json:"message"` // Сообщение
}
