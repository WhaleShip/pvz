package pvz_errors

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

var (
	// user
	ErrUserNotFound      = errors.New("пользователь не найден")
	ErrUserAlreadyExists = errors.New("пользователь с таким email существует")
	ErrInvalidPassword   = errors.New("неверный пароль")
	ErrInvalidRole       = errors.New("некорректная роль")

	// pvz
	ErrInsertPVZFailed     = errors.New("ошибка добавления ПВЗ")
	ErrInvalidPVZCity      = errors.New("некорректный город")
	ErrPVZNotFound         = errors.New("ПВЗ не найден")
	ErrNoOpenRecetionOrPvz = errors.New("нет открытых приёмок или ПВЗ")

	// receptions
	ErrOpenReceptionExists  = errors.New("открытая приёмка существует")
	ErrCloseReceptionFailed = errors.New("ПВЗ или приёмка не найдена")

	// products
	ErrInvalidProduct  = errors.New("некорректный тип продукта")
	ErrDeletingProduct = errors.New("не удалось удалить продукт")

	// middlewares
	ErrMissingAuthHeader       = errors.New("отсутствует заголовок авторизации")
	ErrInvalidAuthHeader       = errors.New("некорректный заголовок авторизации")
	ErrInvalidToken            = errors.New("некорректный токен: ")
	ErrInsufficientPermissions = errors.New("недостаточно прав")
)

func GetErrorStatusCode(err error) int {
	switch {
	// user
	case errors.Is(err, ErrUserAlreadyExists):
		return fiber.StatusConflict
	case errors.Is(err, ErrInvalidRole):
		return fiber.StatusBadRequest
	case errors.Is(err, ErrUserNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, ErrInvalidPassword):
		return fiber.StatusUnauthorized

	// pvz
	case errors.Is(err, ErrPVZNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, ErrInvalidPVZCity):
		return fiber.StatusBadRequest
	case errors.Is(err, ErrNoOpenRecetionOrPvz):
		return fiber.StatusBadRequest

	// receptions
	case errors.Is(err, ErrOpenReceptionExists):
		return fiber.StatusConflict
	case errors.Is(err, ErrCloseReceptionFailed):
		return fiber.StatusBadRequest

		// products
	case errors.Is(err, ErrInvalidProduct):
		return fiber.StatusBadRequest
	case errors.Is(err, ErrDeletingProduct):
		return fiber.StatusBadRequest

	default:
		return fiber.StatusInternalServerError
	}
}
