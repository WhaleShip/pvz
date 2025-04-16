package pvz_errors

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrUserNotFound      = errors.New("пользователь не найден")
	ErrUserAlreadyExists = errors.New("пользователь с таким email существует")
	ErrInvalidPassword   = errors.New("неверный пароль")
	ErrInvalidRole       = errors.New("некорректная роль")
	ErrInsertPVZFailed   = errors.New("ошибка добавления ПВЗ")
	ErrInvalidPVZCity    = errors.New("некорректный город")

	ErrMissingAuthHeader       = errors.New("отсутствует заголовок авторизации")
	ErrInvalidAuthHeader       = errors.New("некорректный заголовок авторизации")
	ErrInvalidToken            = errors.New("некорректный токен: ")
	ErrInsufficientPermissions = errors.New("недостаточно прав")
)

func GetErrorStatusCode(err error) int {
	switch {
	case errors.Is(err, ErrUserAlreadyExists):
		return fiber.StatusConflict
	case errors.Is(err, ErrInvalidRole):
		return fiber.StatusBadRequest
	case errors.Is(err, ErrUserNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, ErrInvalidPassword):
		return fiber.StatusUnauthorized
	case errors.Is(err, ErrInvalidPVZCity):
		return fiber.StatusBadRequest

	default:
		return fiber.StatusInternalServerError
	}
}
