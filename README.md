🔗 URL Shortener

Простой и быстрый сервис для сокращения длинных ссылок с веб‑интерфейсом и REST API.

— Короткие ссылки вида `http://localhost:8080/abc123`
— Автоматический редирект и подсчёт кликов
— Повторное использование кода для одинаковых URL (без дублей)
— Встроенная страница `web/` для ручного тестирования


Что внутри

- Backend: Go 1.21+, `github.com/gorilla/mux`
- Генерация коротких кодов: криптографически стойкая (`pkg/shortener`)
- Веб‑интерфейс: чистый HTML/CSS/JS (`web/`)
- Контейнеризация: Dockerfile, docker‑compose


Архитектура (папки)

- `cmd/main.go` — автономный HTTP‑сервер с обработчиками и доступом к БД напрямую
- `internal/` — «слоистая» архитектура для более крупного приложения:
  - `internal/handler` — HTTP‑обработчики поверх сервиса
  - `internal/service` — бизнес‑логика, валидация, счётчик кликов
  - `internal/repository` — доступ к БД (интерфейс и реализация)
  - `internal/models` — модели запросов/ответов и сущностей
- `pkg/shortener` — генерация коротких кодов фиксированной длины
- `web/` — фронтенд: форма сокращения, просмотр информации, тест редиректа

Сценарий A: Локальный запуск с MySQL (совместим с `cmd/main.go`)

1) Поднимите MySQL 8.0 локально (например, Docker):

   ```bash
   docker run --name url-mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=url_shortener -e MYSQL_USER=urluser -e MYSQL_PASSWORD=password -p 3306:3306 -d mysql:8
   ```

2) Проверьте и при необходимости скорректируйте DSN в `cmd/main.go` (по умолчанию):

   ```
   urluser:password@tcp(localhost:3306)/url_shortener?parseTime=true
   ```

3) Сборка и запуск:

   ```bash
   go mod download
   go run ./cmd
   ```

4) Откройте веб‑интерфейс: `http://localhost:8080`


API

- POST `/api/v1/shorten`
  - Тело: `{ "url": "https://example.com" }`
  - Ответ: `201` `{ "short_url": "abc123" }` (или уже существующий код для дубликатов)

- GET `/api/v1/url/{short}`
  - Ответ: `200` с данными ссылки, например:
    ```json
    {
      "id": "abc123",
      "original_url": "https://example.com",
      "short_url": "abc123",
      "created_at": "2025-10-15T12:00:00Z",
      "clicks": 3
    }
    ```
  - `404`, если не найдено

- GET `/{short}`
  - 302/Found редирект на оригинальный URL, параллельно увеличивается счётчик кликов

- GET `/health`
  - `200 OK` — сервис жив


Веб‑интерфейс (`web/`)

Открывается на корне `http://localhost:8080` и использует API:

- Создание короткой ссылки (форма «Создать короткую ссылку»)
- Получение информации по короткому коду
- Тестирование редиректа (без перехода, через ручной fetch с `redirect: 'manual'`)
- Мониторинг статуса API/БД (раздел «Статус системы»)


Сборка в Docker

```bash
cd URLcutter
docker build -t urlcutter-app .
docker run --rm -p 8080:8080 --name urlcutter-app urlcutter-app
```

Тесты

- Юнит‑тесты для сервиса/хендлеров находятся в `internal/service/service_test.go`, `internal/handler/handler_test.go`.
- Запуск тестов:

  ```bash
  go test ./...
  ```

Лицензия

См. `LICENSE` в корне проекта.
