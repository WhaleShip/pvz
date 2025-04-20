# Сервис для работы с ПВЗ

[![Coverage Status](https://coveralls.io/repos/github/WhaleShip/pvz/badge.svg?branch=main)](https://coveralls.io/github/WhaleShip/pvz?branch=main)
[![linters](https://github.com/WhaleShip/pvz/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/WhaleShip/pvz/actions/workflows/golangci-lint.yml)
[![unit-tests](https://github.com/WhaleShip/pvz/actions/workflows/unit-tests.yml/badge.svg)](https://github.com/WhaleShip/pvz/actions/workflows/unit-tests.yml)

## Краткая сводка

- ### написано на go 1.24.2 с использованием fiber (prefork)
- ### Для регистрации используется stateless JWT
- ### в качестве бд используется postgresql с pgBouncer
- ### интегарционныые тесты находятся в папке tests/integreation_test 
- ### [конфигурация линтера](.golangci.yaml)

> [!IMPORTANT]  
> Дисклеймер: Разработка велась под Linux, у некоторых команд могут возникать проблемы с запуском на Windows, написал решение для всех изввестных мне проблем, но не могу быть уверен.

# Запуск

### 1. Создать .env
> на windows работает только из gitbash

```sh
make env 
```

или переименовать [examplse.env](example.env) в .env


### 2. Запустить через докер
```sh
make run
```

или

```sh
docker compose up
```

у виндовс могут быть проблемы со скриптом баунсера, если такое происходит
```sh
dos2unix docker/scripts/entrypoint.sh
```

основное приложение: http://localhost:8080

grpc доступно на http://localhost:3000

метрики отдаются на http://localhost:9000, графический интерфейс доступен на http://localhost:9090



## Остальной функционал
### unit тесты
> только linux
- Запустить тесты
```sh
make test
```

- Посмотреть покрытие
```sh
make cover # через консоль
make cover-html # через html файл
```


### Интеграционные тесты
> для выполнения нужен докер, так что запускать нужно со среды где есть docker
```sh
make test-int
```

### контакты
[![Telegram Icon](https://raw.githubusercontent.com/CLorant/readme-social-icons/main/large/light/telegram.svg)](https://t.me/PanHater)
[![medium-light-discord](https://raw.githubusercontent.com/CLorant/readme-social-icons/main/large/light/discord.svg)](https://discord.com/users/1249015796852719617)