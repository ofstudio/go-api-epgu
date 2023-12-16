package apipgu

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ofstudio/go-api-epgu/dto"
)

// # Ошибки первого уровня
var (
	ErrOrderCreate        = errors.New("ошибка OrderCreate")
	ErrPushChunked        = errors.New("ошибка OrderPushChunked")
	ErrPush               = errors.New("ошибка OrderPush")
	ErrOrderInfo          = errors.New("ошибка OrderInfo ")
	ErrOrderCancel        = errors.New("ошибка OrderCancel")
	ErrAttachmentDownload = errors.New("ошибка AttachmentDownload")
	ErrService            = errors.New("ошибка услуги")
)

// # Ошибки второго уровня
var (
	ErrMultipartBody         = errors.New("ошибка подготовки multipart-содержимого")
	ErrRequest               = errors.New("ошибка HTTP-запроса")
	ErrUnexpectedContentType = errors.New("неожиданный тип содержимого")
	ErrJSONUnmarshal         = errors.New("ошибка чтения JSON")
	ErrNoFiles               = errors.New("нет файлов во вложении")
	ErrZip                   = errors.New("ошибка создания zip-архива")
	ErrGUID                  = errors.New("не удалось сгенерировать GUID")
	ErrXMLMarshal            = errors.New("ошибка создания XML")
	ErrNilArchive            = errors.New("не передан архив")
	ErrWrongOrderID          = errors.New("некорректный ID заявления")
	ErrInvalidFileLink       = errors.New("некорректная ссылка на файл")
)

// # HTTP-ошибки
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.12",
// "Приложение 4. Ошибки, возвращаемые при запросах к API ЕПГУ"
var (
	ErrStatusOrderNotFound         = errors.New("заявление не найдено")         // HTTP 204
	ErrStatusBadRequest            = errors.New("неверные параметры")           // HTTP 400
	ErrStatusUnauthorized          = errors.New("отказ в доступе")              // HTTP 401
	ErrStatusForbidden             = errors.New("доступ запрещен")              // HTTP 403
	ErrStatusURLNotFound           = errors.New("не найден URL запроса")        // HTTP 404
	ErrStatusUnableToHandleRequest = errors.New("невозможно обработать запрос") // HTTP 409
	ErrStatusTooManyRequests       = errors.New("слишком много запросов")       // HTTP 429
	ErrStatusInternalError         = errors.New("внутренняя ошибка")            // HTTP 500
	ErrStatusBadGateway            = errors.New("некорректный шлюз")            // HTTP 502
	ErrStatusServiceUnavailable    = errors.New("сервис недоступен")            // HTTP 503
	ErrStatusGatewayTimeout        = errors.New("шлюз не отвечает")             // HTTP 504
	ErrStatusUnexpected            = errors.New("неожиданный HTTP-статус")      // Другие HTTP-коды ошибок
)

// # Ошибки ЕПГУ
//
// Подробнее см. "Спецификация API ЕПГУ версия 1.12",
// "Приложение 4. Ошибки, возвращаемые при запросах к API ЕПГУ"
//
// Пример JSON-ответа от ЕПГУ при ошибке:
//
//	{
//	  "code": "order_access",
//	  "message": "У пользователя нет прав для работы с текущим заявлением"
//	}
var (

	// Ошибка ЕПГУ: code = access_denied_person_permissions
	ErrCodeAccessDeniedPersonPermissions = errors.New("пользователь не дал согласие Вашей системе на выполнение данной операции")

	// Ошибка ЕПГУ: code = access_denied_service
	ErrCodeAccessDeniedService = errors.New("доступ ВИС к запрашиваемой услуге запрещен")

	// Ошибка ЕПГУ: code = access_denied_system
	ErrCodeAccessDeniedSystem = errors.New("доступ запрещен для ВИС, отправляющей запрос")

	// Ошибка ЕПГУ: code = access_denied_user
	ErrCodeAccessDeniedUser = errors.New("доступ запрещен для данного типа пользователя")

	// Ошибка ЕПГУ: code = access_denied_user_legal
	ErrCodeAccessDeniedUserLegal = errors.New("попытка создать заявления с использованием токена, полученного для организации, которая не является владельцем ВИС, отправляющей данный запрос")

	// Ошибка ЕПГУ: code = bad_delegation
	ErrCodeBadDelegation = errors.New("нет необходимых полномочий для создания заявления")

	// Ошибка ЕПГУ: code = bad_request
	ErrCodeBadRequest = errors.New("ошибка в параметрах запроса")

	// Ошибка ЕПГУ: code = cancel_not_allowed
	ErrCodeCancelNotAllowed = errors.New("отмена заявления в текущем статусе невозможна")

	// Ошибка ЕПГУ: code = config_delegation
	ErrCodeConfigDelegation = errors.New("полномочие для создания и подачи заявления по заданной услуги не существует")

	// Ошибка ЕПГУ: code = internal_error
	ErrCodeInternalError = errors.New("ошибка в обработке заявления, причины которой можно выяснить при анализе инцидента")

	// Ошибка ЕПГУ: code = limitation_exception
	ErrCodeLimitationException = errors.New("превышение установленных ограничений, указанных в Приложении 3 к Спецификации")

	// Ошибка ЕПГУ: code = not_found
	ErrCodeNotFound = errors.New("заявление не найдено")

	// Ошибка ЕПГУ: code = order_access
	ErrCodeOrderAccess = errors.New("у пользователя нет прав для работы с текущим заявлением")

	// Ошибка ЕПГУ: code = push_denied
	ErrCodePushDenied = errors.New("нет прав для отправки заявления. Отправить заявление может только руководитель организации или сотрудник с доверенностью")

	// Ошибка ЕПГУ: code = service_not_found
	ErrCodeServiceNotFound = errors.New("не найдена услуга, заданная кодом serviceCode в запросе")

	// Ошибка ЕПГУ: неизвестное значение code
	ErrCodeUnexpected = errors.New("неожиданный код ошибки")

	// Ошибка ЕПГУ: code не указан
	ErrCodeNotSpecified = errors.New("код ошибки не указан")
)

// HTTP 403 Forbidden: доступ запрещен: доступ запрещен для ВИС, отправляющей запрос [code='access_denied_system', message='ValidationCommonError.notAllowed']
func responseError(res *http.Response) error {
	if res == nil || (res.StatusCode != 204 && res.StatusCode < 400) {
		return nil
	}

	if res.StatusCode == 204 {
		return fmt.Errorf("HTTP %s: %w", res.Status, ErrStatusOrderNotFound)
	}

	return fmt.Errorf(
		"HTTP %s: %w: %w",
		res.Status, httpStatusError(res.StatusCode), bodyError(res),
	)
}

func httpStatusError(statusCode int) error {
	switch statusCode {
	case 400:
		return ErrStatusBadRequest
	case 401:
		return ErrStatusUnauthorized
	case 403:
		return ErrStatusForbidden
	case 404:
		return ErrStatusURLNotFound
	case 409:
		return ErrStatusUnableToHandleRequest
	case 429:
		return ErrStatusTooManyRequests
	case 500:
		return ErrStatusInternalError
	case 502:
		return ErrStatusBadGateway
	case 503:
		return ErrStatusServiceUnavailable
	case 504:
		return ErrStatusGatewayTimeout
	default:
		return ErrStatusUnexpected
	}
}

func bodyError(res *http.Response) error {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
	}
	ct := res.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		return jsonError(body)
	}
	return fmt.Errorf("%w: '%s'", ErrUnexpectedContentType, ct)
}

func jsonError(body []byte) error {
	errResponse := &dto.ErrorResponse{}
	err := json.Unmarshal(body, errResponse)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrJSONUnmarshal, err)
	}

	switch errResponse.Code {
	case "access_denied_person_permissions":
		err = ErrCodeAccessDeniedPersonPermissions
	case "access_denied_service":
		err = ErrCodeAccessDeniedService
	case "access_denied_system":
		err = ErrCodeAccessDeniedSystem
	case "access_denied_user":
		err = ErrCodeAccessDeniedUser
	case "access_denied_user_legal":
		err = ErrCodeAccessDeniedUserLegal
	case "bad_delegation":
		err = ErrCodeBadDelegation
	case "bad_request":
		err = ErrCodeBadRequest
	case "cancel_not_allowed":
		err = ErrCodeCancelNotAllowed
	case "config_delegation":
		err = ErrCodeConfigDelegation
	case "internal_error":
		err = ErrCodeInternalError
	case "limitation_exception":
		err = ErrCodeLimitationException
	case "not_found":
		err = ErrCodeNotFound
	case "order_access":
		err = ErrCodeOrderAccess
	case "push_denied":
		err = ErrCodePushDenied
	case "service_not_found":
		err = ErrCodeServiceNotFound
	case "":
		err = ErrCodeNotSpecified
	default:
		err = ErrCodeUnexpected
	}

	var fields []string
	if errResponse.Code != "" {
		fields = append(fields, fmt.Sprintf("code='%s'", errResponse.Code))
	}
	if errResponse.Message != "" {
		fields = append(fields, fmt.Sprintf("message='%s'", errResponse.Message))
	}
	if errResponse.Error != "" {
		fields = append(fields, fmt.Sprintf("error='%s'", errResponse.Error))
	}

	return fmt.Errorf("%w [%s]", err, strings.Join(fields, ", "))
}
