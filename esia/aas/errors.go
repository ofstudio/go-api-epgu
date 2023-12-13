package aas

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// # Ошибки первого уровня
var (
	ErrAuthURI       = errors.New("ошибка при создании авторизационной ссылки")
	ErrParseCallback = errors.New("ошибка обратного вызова")
	ErrTokenExchange = errors.New("ошибка запроса токена")
	ErrTokenUpdate   = errors.New("ошибка обновления токена")
)

// # Ошибки второго уровня
var (
	ErrNoState               = errors.New("отсутствует поле state")
	ErrGUID                  = errors.New("не удалось сгенерировать GUID")
	ErrSign                  = errors.New("ошибка подписания")
	ErrRequest               = errors.New("ошибка HTTP-запроса")
	ErrJSONUnmarshal         = errors.New("ошибка чтения JSON")
	ErrUnexpectedContentType = errors.New("неожиданный тип содержимого")
)

// # Ошибки ЕСИА
var (
	ErrESIA_036700  = errors.New("ESIA-036700: Не указана мнемоника типа согласия")
	ErrESIA_036701  = errors.New("ESIA-036701: Не найден тип согласия")
	ErrESIA_036702  = errors.New("ESIA-036702: Не указан обязательный скоуп для типа согласия")
	ErrESIA_036703  = errors.New("ESIA-036703: Указанные скоупы выходят за рамки разрешенных для типа согласия")
	ErrESIA_036704  = errors.New("ESIA-036704: Запрещено указывать скоупы для типа согласия")
	ErrESIA_036705  = errors.New("ESIA-036705: Необходимо указать хотя бы одно действие")
	ErrESIA_036706  = errors.New("ESIA-036706: Указанное действие не существует")
	ErrESIA_036707  = errors.New("ESIA-036707: Необходимо указать хотя бы одну цель")
	ErrESIA_036716  = errors.New("ESIA-036716: Указано некорректное время истечения срока действия согласия")
	ErrESIA_036726  = errors.New("ESIA-036726: Указанная цель не существует")
	ErrESIA_036727  = errors.New("ESIA-036727: Необходимо указать одну цель согласия")
	ErrESIA_007002  = errors.New("ESIA-007002: Несоответствие сертификата и мнемоники информационной системы или отсутствие сертификата для данной системы в ЕСИА")
	ErrESIA_007003  = errors.New("ESIA-007003: В запросе отсутствует обязательный параметр, запрос включает в себя неверное значение параметра \nили включает параметр несколько раз\n")
	ErrESIA_007004  = errors.New("ESIA-007004: Владелец ресурса или сервис авторизации отклонил запрос")
	ErrESIA_007005  = errors.New("ESIA-007005: Система-клиент не имеет права запрашивать получение маркера доступа таким методом")
	ErrESIA_007006  = errors.New("ESIA-007006: Запрошенная область доступа (scope) указана неверно, неизвестно или сформирована некорректно")
	ErrESIA_007007  = errors.New("ESIA-007007: Возникла неожиданная ошибка в работе сервиса авторизации, которая привела к невозможности выполнить запрос")
	ErrESIA_007008  = errors.New("ESIA-007008: Сервис авторизации в настоящее время не может выполнить запрос из-за большой нагрузки или технических работ на сервере")
	ErrESIA_007009  = errors.New("ESIA-007009: Сервис авторизации не поддерживает получение маркера доступа этим методом")
	ErrESIA_007011  = errors.New("ESIA-007011: Авторизационный код или маркер обновления недействителен, просрочен, отозван или не соответствует адресу ресурса, указанному в запросе на авторизацию, или был выдан другой системе-клиенту")
	ErrESIA_007012  = errors.New("ESIA-007012: Тип авторизационного кода не поддерживается сервисом авторизации")
	ErrESIA_007013  = errors.New("ESIA-007013: Запрос не содержит указания на область доступа (scope)")
	ErrESIA_007014  = errors.New("ESIA-007014: Запрос не содержит обязательного параметра")
	ErrESIA_007015  = errors.New("ESIA-007015: Неверное время запроса")
	ErrESIA_007019  = errors.New("ESIA-007019: Отсутствует разрешение на доступ")
	ErrESIA_007023  = errors.New("ESIA-007023: Указанный в запросе <redirect_uri> отсутствует среди разрешенных для ИС")
	ErrESIA_007038  = errors.New("ESIA-007038: Ошибка получения параметров из запроса")
	ErrESIA_007039  = errors.New("ESIA-007039: В изначальном запросе на /v2/ac, параметр <code_challenge> не был указан")
	ErrESIA_007040  = errors.New("ESIA-007040: Ошибка сравнения исходного и контрольного значений")
	ErrESIA_007046  = errors.New("ESIA-007046: Запрос otp невозможен, а в области доступа (scope) указано обязательное прохождение пользователем двухфакторной авторизации, недоступный пользователю")
	ErrESIA_007053  = errors.New("ESIA-007053: client_secret сформирован некорректно. client_secret не соответствует строке-сертификату, информационной системе или используемый сертификат не активен")
	ErrESIA_007055  = errors.New("ESIA-007055: Вход в систему осуществляется с неподтвержденной учетной записью")
	ErrESIA_007060  = errors.New("ESIA-007060: Значение параметра <roles> в запросе указано неверно")
	ErrESIA_007061  = errors.New("ESIA-007061: Значение параметра <obj_type> в запросе указано неверно")
	ErrESIA_007062  = errors.New("ESIA-007062: Тип или роль пользователя в запросе указана неверно")
	ErrESIA_007194  = errors.New("ESIA-007194: Запрос области доступа для организации, сотрудником которой пользователь не является ")
	ErrESIA_008010  = errors.New("ESIA-008010: Не удалось произвести аутентификацию системы-клиента")
	ErrESIA_unknown = errors.New("неизвестная ошибка ЕСИА")
)

// callbackError - возвращает ошибку ЕСИА по коду ошибки в query-параметрах callback-запроса к redirect_uri от ЕСИА.
// Пример ошибки:
//
//	ESIA-007014: Запрос не содержит обязательного параметра [error='invalid_request', error_description='ESIA-007014: The request does not contain the mandatory parameter' state='48d1a8dc-0b7d-418a-b4ef-2c7797f77dc9']'
func callbackError(query url.Values) error {
	errorRes := ErrorResponse{
		Error:            query.Get("error"),
		ErrorDescription: query.Get("error_description"),
		State:            query.Get("state"),
	}

	var err error
	switch {
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036700"):
		err = ErrESIA_036700
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036701"):
		err = ErrESIA_036701
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036702"):
		err = ErrESIA_036702
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036703"):
		err = ErrESIA_036703
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036704"):
		err = ErrESIA_036704
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036705"):
		err = ErrESIA_036705
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036706"):
		err = ErrESIA_036706
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036707"):
		err = ErrESIA_036707
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036716"):
		err = ErrESIA_036716
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036726"):
		err = ErrESIA_036726
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-036727"):
		err = ErrESIA_036727
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007002"):
		err = ErrESIA_007002
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007003"):
		err = ErrESIA_007003
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007004"):
		err = ErrESIA_007004
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007005"):
		err = ErrESIA_007005
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007006"):
		err = ErrESIA_007006
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007007"):
		err = ErrESIA_007007
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007008"):
		err = ErrESIA_007008
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007009"):
		err = ErrESIA_007009
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-008010"):
		err = ErrESIA_008010
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007011"):
		err = ErrESIA_007011
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007012"):
		err = ErrESIA_007012
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007013"):
		err = ErrESIA_007013
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007014"):
		err = ErrESIA_007014
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007015"):
		err = ErrESIA_007015
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007019"):
		err = ErrESIA_007019
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007023"):
		err = ErrESIA_007023
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007046"):
		err = ErrESIA_007046
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007053"):
		err = ErrESIA_007053
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007055"):
		err = ErrESIA_007055
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007060"):
		err = ErrESIA_007060
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007061"):
		err = ErrESIA_007061
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007062"):
		err = ErrESIA_007062
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007194"):
		err = ErrESIA_007194
	default:
		err = ErrESIA_unknown
	}
	return fmt.Errorf(
		"%w [error='%s', error_description='%s', state='%s']",
		err, errorRes.Error, errorRes.ErrorDescription, errorRes.State,
	)
}

// exchangeError - возвращает ошибку ЕСИА по коду ошибки в ответе от ЕСИА при обмене кода на маркер доступа.
// Пример ошибки:
//
//	HTTP 400 Bad request: ESIA-007014: Запрос не содержит обязательного параметра [error='invalid_request', error_description='ESIA-007014: The request does not contain the mandatory parameter' state='48d1a8dc-0b7d-418a-b4ef-2c7797f77dc9']'
func exchangeError(res *http.Response) error {
	if res == nil || res.StatusCode < 400 {
		return nil
	}
	return fmt.Errorf("HTTP %s: %w", res.Status, bodyError(res))
}

func bodyError(res *http.Response) error {
	//goland:noinspection ALL
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequest, err)
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
	errorRes := &ErrorResponse{}
	err := json.Unmarshal(body, errorRes)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrJSONUnmarshal, err)
	}

	switch {
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007003"):
		err = ErrESIA_007003
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007004"):
		err = ErrESIA_007004
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007005"):
		err = ErrESIA_007005
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007006"):
		err = ErrESIA_007006
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007007"):
		err = ErrESIA_007007
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007008"):
		err = ErrESIA_007008
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007009"):
		err = ErrESIA_007009
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-008010"):
		err = ErrESIA_008010
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007011"):
		err = ErrESIA_007011
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007012"):
		err = ErrESIA_007012
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007013"):
		err = ErrESIA_007013
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007014"):
		err = ErrESIA_007014
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007015"):
		err = ErrESIA_007015
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007019"):
		err = ErrESIA_007019
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007023"):
		err = ErrESIA_007023
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007038"):
		err = ErrESIA_007038
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007039"):
		err = ErrESIA_007039
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007040"):
		err = ErrESIA_007040
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007046"):
		err = ErrESIA_007046
	case strings.HasPrefix(errorRes.ErrorDescription, "ESIA-007053"):
		err = ErrESIA_007053
	default:
		err = ErrESIA_unknown
	}

	return fmt.Errorf(
		"%w [error='%s', error_description='%s', state='%s']",
		err, errorRes.Error, errorRes.ErrorDescription, errorRes.State,
	)
}

func textError(body []byte) error {
	return errors.New(strings.Replace(string(body), "\n", "\\n", -1))
}
