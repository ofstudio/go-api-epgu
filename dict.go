package apipgu

// Типы запрашиваемого справочника (плоский / иерархический)
// для метода [Client.Dict].
const (
	DictFilterOneLevel string = "ONELEVEL" // Плоский справочник
	DictFilterSubTree  string = "SUBTREE"  // Иерархический справочник
)

// DictItem - элемент справочника.
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.12",
// раздел "3. Получение справочных данных".
//
// Пример элемента справочника EXTERNAL_BIC:
//
//	{
//	  "value": "044525974",
//	  "title": "044525974 - АО \"Тинькофф Банк\" г Москва",
//	  "isLeaf": true,
//	  "children": [],
//	  "attributes": [
//	    {
//	      "name": "ID",
//	      "type": "STRING",
//	      "value": {
//	        "asString": "044525974",
//	        "typeOfValue": "STRING",
//	        "value": "044525974"
//	      },
//	      "valueAsOfType": "044525974"
//	    },
//	    {
//	      "name": "NAME",
//	      "type": "STRING",
//	      "value": {
//	        "asString": "АО \"Тинькофф Банк\" г Москва",
//	        "typeOfValue": "STRING",
//	        "value": "АО \"Тинькофф Банк\" г Москва"
//	      },
//	      "valueAsOfType": "АО \"Тинькофф Банк\" г Москва"
//	    },
//	    {
//	      "name": "BIC",
//	      "type": "STRING",
//	      "value": {
//	        "asString": "044525974",
//	        "typeOfValue": "STRING",
//	        "value": "044525974"
//	      },
//	      "valueAsOfType": "044525974"
//	    },
//	    {
//	      "name": "CORR_ACCOUNT",
//	      "type": "STRING",
//	      "value": {
//	        "asString": "30101810145250000974",
//	        "typeOfValue": "STRING",
//	        "value": "30101810145250000974"
//	      },
//	      "valueAsOfType": "30101810145250000974"
//	    }
//	  ],
//	  "attributeValues": {
//	    "ID": "044525974",
//	    "CORR_ACCOUNT": "30101810145250000974",
//	    "BIC": "044525974",
//	    "NAME": "АО \"Тинькофф Банк\" г Москва"
//	  }
//	}
//
// Пример элемента справочника TO_PFR:
//
//	{
//	  "value": "087109",
//	  "title": "Клиентская служба «Замоскворечье, Якиманка» по г. Москве и МО",
//	  "isLeaf": true,
//	  "children": [],
//	  "attributes": [],
//	  "attributeValues": {}
//	},
type DictItem struct {
	Value           string           `json:"value"`                 // Код элемента справочника
	ParentValue     string           `json:"parentValue,omitempty"` // Код родительского элемента
	Title           string           `json:"title"`                 // Наименование элемента
	IsLeaf          bool             `json:"isLeaf"`                // [?] Признак наличия подчинённых элементов
	Children        []map[string]any `json:"children"`              // Подчинённые элементы
	Attributes      []DictAttribute  `json:"attributes"`            // Дополнительные атрибуты элемента справочника [детально]
	AttributeValues map[string]any   `json:"attributeValues"`       // Список значений дополнительных атрибутов элемента справочника [кратко]
}

// DictAttribute - дополнительный атрибут элемента справочника из структуры [DictItem].
type DictAttribute struct {
	Name          string             `json:"name"`          // [Не документировано]
	Type          string             `json:"type"`          // [Не документировано]
	Value         DictAttributeValue `json:"value"`         // [Не документировано]
	ValueAsOfType any                `json:"valueAsOfType"` // [Не документировано]
}

// DictAttributeValue - значение дополнительного атрибута элемента справочника из структуры [DictAttribute].
type DictAttributeValue struct {
	AsString    string `json:"asString"`    // [Не документировано]
	TypeOfValue string `json:"typeOfValue"` // [Не документировано]
	Value       any    `json:"value"`       // [Не документировано]
}
