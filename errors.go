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

var (
	ErrMultipartBodyPrepare  = errors.New("ошибка подготовки multipart-содержимого")
	ErrRequestPrepare        = errors.New("ошибка подготовки запроса")
	ErrRequestCall           = errors.New("ошибка вызова запроса")
	ErrResponseRead          = errors.New("ошибка чтения ответа")
	ErrUnexpectedContentType = errors.New("неожиданный тип содержимого")
	ErrUnmarshal             = errors.New("ошибка чтения JSON")
	ErrNoFiles               = errors.New("нет файлов во вложении")
	ErrZipCreate             = errors.New("ошибка создания файла в zip-архиве")
	ErrZipWrite              = errors.New("ошибка записи файла в zip-архив")
	ErrZipClose              = errors.New("ошибка закрытия zip-архива")

	ErrOrderCreate = errors.New("ошибка OrderCreate")
	ErrPushChunked = errors.New("ошибка PushChunked")
	ErrOrderInfo   = errors.New("ошибка OrderInfo ")
)

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
	ErrStatusStatusUnexpected      = errors.New("неожиданный HTTP-статус")      // other HTTP codes
)

var (
	ErrCodeAccessDeniedPersonPermissions = errors.New("пользователь не дал согласие Вашей системе на выполнение данной операции")
	ErrCodeAccessDeniedService           = errors.New("доступ ВИС к запрашиваемой услуге запрещен")
	ErrCodeAccessDeniedSystem            = errors.New("доступ запрещен для ВИС, отправляющей запрос")
	ErrCodeAccessDeniedUser              = errors.New("доступ запрещен для данного типа пользователя")
	ErrCodeAccessDeniedUserLegal         = errors.New("попытка создать заявления с использованием токена, полученного для организации, которая не является владельцем ВИС, отправляющей данный запрос")
	ErrCodeBadDelegation                 = errors.New("нет необходимых полномочий для создания заявления")
	ErrCodeBadRequest                    = errors.New("ошибка в параметрах запроса")
	ErrCodeCancelNotAllowed              = errors.New("отмена заявления в текущем статусе невозможна")
	ErrCodeConfigDelegation              = errors.New("полномочие для создания и подачи заявления по заданной услуги не существует")
	ErrCodeInternalError                 = errors.New("ошибка в обработке заявления, причины которой можно выяснить при анализе инцидента")
	ErrCodeLimitationException           = errors.New("превышение установленных ограничений, указанных в Приложении 3 к Спецификации")
	ErrCodeNotFound                      = errors.New("заявление не найдено")
	ErrCodeOrderAccess                   = errors.New("у пользователя нет прав для работы с текущим заявлением")
	ErrCodePushDenied                    = errors.New("нет прав для отправки заявления. Отправить заявление может только руководитель организации или сотрудник с доверенностью")
	ErrCodeServiceNotFound               = errors.New("не найдена услуга, заданная кодом serviceCode в запросе")
	ErrCodeUnexpected                    = errors.New("неожиданный код ошибки")
)

// HTTP 403 Forbidden: доступ запрещен: доступ запрещен для ВИС, отправляющей запрос [code='access_denied_system', message='ValidationCommonError.notAllowed']
func responseError(res *http.Response) error {
	if res == nil || (res.StatusCode != 204 && res.StatusCode < 400) {
		return nil
	}
	return fmt.Errorf(
		"HTTP %s: %w: %w",
		res.Status, httpStatusError(res.StatusCode), bodyError(res),
	)
}

func httpStatusError(statusCode int) error {
	switch statusCode {
	case 204:
		return ErrStatusOrderNotFound
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
		return ErrStatusStatusUnexpected
	}
}

func bodyError(res *http.Response) error {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrResponseRead, err)
	}
	ct := res.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(ct, "application/json"):
		return jsonError(body)
	case strings.HasPrefix(ct, "text/plain") || ct == "":
		return textError(body)
	default:
		return fmt.Errorf("%w [%s]", ErrUnexpectedContentType, ct)
	}
}

func jsonError(body []byte) error {
	errorRes := &dto.ErrorResponse{}
	err := json.Unmarshal(body, errorRes)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnmarshal, err)
	}

	switch errorRes.Code {
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
	default:
		err = ErrCodeUnexpected
	}

	return fmt.Errorf(" %w [code='%s', message='%s']", err, errorRes.Code, errorRes.Message)
}

func textError(body []byte) error {
	return errors.New(strings.Replace(string(body), "\n", "\\n", -1))
}
