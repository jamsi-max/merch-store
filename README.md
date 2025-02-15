# 🛍 Магазин мерча для сотрудников

## 📝 Описание проекта

Магазин мерча — это сервис, который позволяет сотрудникам компании обмениваться виртуальными монетами и приобретать на них мерч.
Каждый сотрудник может:

- Просматривать список купленных товаров.
- Отслеживать историю переводов монет, включая информацию о том, кто и в каком количестве передавал монеты.
- Совершать переводы монет другим пользователям.
- Покупать мерч за накопленные монеты.

Монетный баланс не может быть отрицательным: система предотвращает уход в минус при операциях с монетами.

## 🛠 Используемые технологии

- **Golang** 1.23
- **Gin** (веб-фреймворк)
- **sqlx** (расширенная работа с SQL)
- **Goose** (миграции базы данных)
- **PostgreSQL** 17.4
- **Docker Compose** (для управления контейнерами)

## 📡 API эндпоинты

### 1. Получение информации о балансе

**GET** `/api/info`

**Описание:** Возвращает информацию о текущем балансе монет, купленных товарах и истории транзакций пользователя.

**Пример запроса:**

```sh
curl -H "Authorization: Bearer <TOKEN>" -X GET http://localhost:8080/api/info
```

**Пример успешного ответа `200 OK`**

```json
{
  "coins": 1000,
  "inventory": [
    {
      "type": "powerbank",
      "quantity": 1
    }
  ],
  "coinHistory": {
    "received": [
      {
        "fromUser": "john_doe",
        "amount": 50
      }
    ],
    "sent": [
      {
        "toUser": "jane_doe",
        "amount": 30
      }
    ]
  }
}
```

**Пример ответа с ошибкой (400, 401, 500):**

```json
{
  "errors": "Описание ошибки"
}
```

### 2. Отправка монет

**POST** `/api/sendCoin`

**Описание:** Позволяет отправить монеты другому пользователю.

**Пример запроса:**

```sh
curl -H "Authorization: Bearer <TOKEN>" \
     -H "Content-Type: application/json" \
     -X POST http://localhost:8080/api/sendCoin \
     -d '{
       "toUser": "john_doe",
       "amount": 50
     }'
```

**Пример успешного ответа `200 OK`**

**Пример ответа с ошибкой (400, 401, 500):**

```json
{
  "errors": "Описание ошибки"
}
```

### 3. Покупка товара

**GET** `/api/buy/{item}`

**Описание:** Позволяет купить товар за монеты.

**Пример запроса:**

```sh
curl -H "Authorization: Bearer <TOKEN>" -X GET http://localhost:8080/api/buy/powerbank
```

**Пример успешного ответа `200 OK`**

**Пример ответа с ошибкой (400, 401, 500):**

```json
{
  "errors": "Описание ошибки"
}
```

### 4. Аутентификация

**POST** `/api/auth`

**Описание:** Позволяет аутентифицировать пользователя и получить JWT-токен.

**Пример запроса:**

```sh
curl -H "Content-Type: application/json" \
     -X POST http://localhost:8080/api/auth \
     -d '{
       "username": "user",
       "password": "pass"
     }'
```

**Пример успешного ответа `200 OK`**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Пример ответа с ошибкой (400, 401, 500):**

```json
{
  "errors": "Описание ошибки"
}
```

## 🚀 Запуск проекта

### Клонирование репозитория

```sh
 git clone <URL_РЕПОЗИТОРИЯ>
```

### Создайте и настройте файл `.env`

В проекте используются следующие переменные окружения:

```sh
DB_CONTAINER_NAME=
DB_USER=
DB_PASS=
DB_NAME=
DB_HOST=
DB_PORT=
DATABASE_URL=
DATABASE_URL_CONTAINER=
APP_CONTAINER_NAME=
JWT_SECRET=
```

### Убедитесь, что у вас установлен Docker Compose

- Установка Docker Compose: [🔗 Официальный сайт](https://docs.docker.com/compose/install/)

### ▶️ Запуск сервиса

```sh
docker compose up -d  # Запуск в фоновом режиме
```

или

```sh
docker compose up  # Запуск с выводом логов
```

### Запуск тестов

```sh
go test ./...
```

### Запуск покрытия тестами

```sh
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep total
```

> **Покрытие > 40%**:
> ![coverage](https://github.com/jamsi-max/merch-store/blob/main/screenshots/coverage.png)

### Запуск линтера (golangci-lint)

- Установка линтера: [🔗 Официальный сайт](https://golangci-lint.run/welcome/install/)

```sh
golangci-lint run
```

## Результаты нагрузочного тестирования

Тестирование проводилось с использованием **k6**.

### Инструмент k6

- k6 — инструмент для нагрузочного тестирования API.
- Установка: [🔗 Официальный сайт](https://k6.io/docs/get-started/installation/)

## ⚠️ Проблемные вопросы, с которыми столкнулся

- Вот улучшенная версия текста:

Наибольшим вызовом в проекте стало обеспечение высокой производительности с заданными метриками: **1000 запросов в секунду (RPS) при времени отклика (SLI) не более 50 миллисекунд**. К сожалению, возможности провести полномасштабное нагрузочное тестирование не было, и я надеюсь получить опыт в этой области во время стажировки.
Учитывая использование транзакций в проекте, я не успел завершить все запланированные оптимизации. В планах было внедрение кэширования с использованием Redis и асинхронная обработка операций – например, немедленный ответ пользователю с последующей асинхронной записью в базу данных. Хотя идей по улучшению производительности было много, временные ограничения не позволили реализовать их в полном объеме..

- Вот улучшенная версия текста:

В процессе разработки пришлось принять решения по обработке различных краевых случаев, которые не были явно описаны в условиях:

1. Обработка неразрешенных HTTP-методов для эндпоинтов
2. Валидация транзакций с отрицательным количеством монет
3. Управление JWT-токенами:
   - Обработка ситуаций с параллельно существующими токенами одного пользователя
   - Действия с валидными токенами удалённых пользователей
   - и др.

Хотя все эти сценарии были реализованы, не уверен, что все правильно.
