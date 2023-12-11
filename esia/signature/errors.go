package signature

import "errors"

// Ошибки провайдера [LocalCryptoPro]
var (
	ErrTempFileCreate = errors.New("ошибка при создании временного файла")
	ErrTempFileWrite  = errors.New("ошибка записи во временный файл")
	ErrTempFileRead   = errors.New("ошибка чтения временного файла")
	ErrCPTestExec     = errors.New("ошибка запуска cptest")
)
