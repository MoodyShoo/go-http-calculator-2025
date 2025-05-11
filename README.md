
# go-http-calculator

Yandex Golang Practice

## HTTP API Калькулятор

### Оглавление

- [Возможности](#возможности)
- [API](#api)
  - [Регистрация](#регистрация)
  - [Авторизация](#авторизация)
  - [Вычисление выражения](#вычисление-выражения)
  - [Список выражений](#список-выражений)
  - [Получение выражения по его ID](#получение-выражения-по-его-id)
- [Установка и настройка](#установка-и-настройка)
- [Тестирование](#тестирование)
- [Как это работает](#как-это-работает)
  - [Сервер (Оркестратор)](#сервер-оркестратор)
    - [Принцип работы /api/v1/calculate](#принцип-работы-apiv1calculate)
    - [Принцип работы /api/v1/expressions](#принцип-работы-apiv1expressions)
    - [Принцип работы /api/v1/expressions/{id}](#принцип-работы-apiv1expressionsid)
    - [Принцип работы /internal/task](#принцип-работы-internaltask)
  - [Агент](#агент)
    - [Получение задачи (fetchTask)](#получение-задачи-fetchtask)
    - [Выполнение задачи (executeTask)](#выполнение-задачи-executetask)
    - [Отправка результата (sendResult)](#отправка-результата-sendresult)
    - [Запуск воркеров (RunGoroutines)](#запуск-воркеров-rungoroutines)
  - [Front-end](#front-end)

## Возможности

- Базовые арифметические операции (`+`, `-`, `*`, `/`)
- Поддержка десятичных чисел (например, `3.14`)
- Учитывает приоритет операций (скобки, умножение, деление)
- Логирование запросов, результатов и ошибок

## API

### Регистрация

**Endpoint:** `POST /api/v1/register`

**Тело запроса:** `Content-Type: application/json`

#### Примеры запросов и ответов

**Запрос:**

```json
{
  "login": "test_user", "password": "qwerty"
}
```

**Ответ (Status 200 OK):**

```json
// Пустое тело
```

---

**Запрос (пользователь уже есть):**

```json
{
  "login": "test_user", "password": "qwerty"
}
```

**Ответ (Status 400 Bad Request):**

```json
{
  "error": "user already exists"
}
```

---

**Запрос (пустые поля):**

```json
{
  "login": "", "password": ""
}
```

**Ответ (Status 401 Unauthorized):**

```json
{
  "error": "login or password can't be empty"
}
```

---

### Авторизация

**Endpoint:** `POST /api/v1/login`

**Тело запроса:** `Content-Type: application/json`

#### Примеры запросов и ответов

**Запрос (пользователь есть):**

```json
{
  "login": "test_user", "password": "qwerty"
}
```

**Ответ (Status 200 OK):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY5OTA3MzMsImlhdCI6MTc0Njk5MDYxMywibmFtZSI6MSwibmJmIjoxNzQ2OTkwNjEzfQ.50OU0cLlqjszPRSA9jeXBA_FIgQNK6R8LbFvYe9lvWU"
}
```

---

**Запрос (пользователя нет):**

```json
{
  "login": "test", "password": "qwerty"
}
```

**Ответ (Status 401 Unauthorized):**

```json
{
  "error": "user not found"
}
```

---

**Запрос (неверный пароль):**

```json
{
  "login": "test_user", "password": "maybe_this"
}
```

**Ответ (Status 401 Unauthorized):**

```json
{
  "error": "invalid password"
}
```

---

### Вычисление выражения

**Endpoint:** `POST /api/v1/calculate`

**Тело запроса:** `Content-Type: application/json`

**В заголовке обязательно должен быть:** `Bearer <TOKEN>`

#### Примеры запросов и ответов

**Запрос:**

```json
{
  "expression": "2+2"
}
```

**Ответ (Status 202 Accepted):**

```json
{
  "id": 1
}
```

---

**Запрос с ошибкой:**

```json
{
  "expression": "2+"
}
```

**Ответ (Status 422 Unprocessable Entity):**

```json
{
  "error": "unprocessable entity"
}
```

---

**Запрос с ошибкой выражения:**

```json
{
  "expression": "2-2+"
}
```

**Ответ (Status 500 Internal Server Error):**

```json
{
  "error": "failed to create tasks: not enough operands for operator: +"
}
```

---

### Список выражений

**Endpoint:** `GET /api/v1/expressions`

**В заголовке обязательно должен быть:** `Bearer <TOKEN>`

#### Примеры ответов на запрос

```json
{
    "expressions": [
        {
            "id": 1,
            "expression": "2+2",
            "status": "done",
            "result": 4
        },
        {
            "id": 2,
            "expression": "2/0",
            "status": "error",
            "result": 0,
            "error": "division by zero"
        },
        {
            "id": 3,
            "expression": "(3 + 5) * (2 - 6) / (4 + 7) * (8 - 3) + (10 / (2 + 3)) - (4 * (5 - 2)) + (12 / (3 + 1)) * 2 ",
            "status": "computing",
            "result": 0
        }
    ]
}
```

У выражений есть несколько статусов:

- pending - в очереди на вычисление
- computing - вычисляется в данный момент
- done - успешно вычисленно
- error - во время вычисления произошла ошибка(если некорректное выражение)

---

### Получение выражения по его ID

**Endpoint:** `GET /api/v1/expressions/{id}`

**В заголовке обязательно должен быть:** `Bearer <TOKEN>`

#### Примеры ответов на запрос

**Некорректный ID:**

```text
/api/v1/expressions/-2
```

**Ответ:**

```json
{
  "error": "expression not found"
}
```

---

**Корректный ID:**

```text
/api/v1/expressions/2
```

**Ответ:**

```json
{
  "id": 2,
  "expression": "(3 + 5) * (2 - 6) / (4 + 7) * (8 - 3) + (10 / (2 + 3)) - (4 * (5 - 2)) + (12 / (3 + 1)) * 2 ",
  "status": "done",
  "result": -18.545455
}
```

## Установка и настройка

1. Клонировать репозиторий с помощью `git clone`:

```bash
# ssh
git clone git@github.com:MoodyShoo/go-http-calculator-2025.git
# https 
https://github.com/MoodyShoo/go-http-calculator-2025.git
```

2. Перейти в репозиторий:

```bash
cd go-http-calculator-2025
```

3. Установить все зависимости

```bash
go mod tidy
```

4. По умолчанию HTTP сервер запускается на порту 8080.
   - Изменить на Windows:

     ```cmd
     set PORT=3000
     ```

     или

     ```powershell
     $env:PORT=3000;
     ```

   - Изменить в Linux:

      ```bash
       PORT=1234
      ```

    Для сервера существуют дополнительные параметры
    - TIME_ADDITION_MS - время выполнения сложения
    - TIME_SUBTRACTION_MS - время выполнения вычитания
    - TIME_MULTIPLICATIONS_MS - время выполнения умножения
    - TIME_DIVISIONS_MS - время выполнения деления

    По умолчанию значения всех параметров равно 1000 millisec.

    ---

    Для агента существуют дополнительные параметры
    - ORCHESTARTOR_ADDRESS - адрес сервера gRPC (По умолчанию localhost)
    - ORCHESTARTOR_PORT - адрес порта gRPC (по умолчанию 5000)
    - COMPUTING_POWER - вол-во воркеров (По умолчанию 2)

5. Запустить сервер:

```text
go run cmd/server/main.go
```

Запустить агента:

```text
go run cmd/agent/main.go
```

6. Остановить приложение:
   Сочетание клавиш `Ctrl + C`

## Пример использования

-Windows

```cmd
curl -X POST http://127.0.0.1:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <твой_токен>" \
  -d "{\"expression\": \"2 + 2\"}"
```

-Linux

```bash
curl -X POST http://127.0.0.1:8080/api/v1/calculate \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <твой_токен>" \
-d '{"expression": "2 + 2"}'
```

Ожидаемый результат:

```json
{"id": </id зависит от кол-ва выражений добавленных до>}
```

# Тестирование

Тестами покрыты:

- Оркестратор - `/internal/orchestrator/orchestrator_test.go`
- Хранилище токенов `/internal/auth/auth_test.go`
- Алгоритм Shunting Yard - `/pkg/calculation/calculation_test.go`

- Запуск тестов
  - Перейти в директорию:

    ```bash
    cd <имя директории выше>
    ```

  - Запустить тесты:

    ```bash
    go test -v . 
    ```

  - Доступные тесты для оркестратора:
    - TestCalculateRoute
      - Valid_expression
      - Invalid_JSON
      - Invalid_expression
      - Invalid_expression
      - Malformed JSON
      - Missing_expression_field
      - Large_expression

    - TestExpressionsHandler
      - One_valid_expression
      - Multiple_valid_expressions
      - Empty_list

    - TestExpressionIdHandler
      - Valid_expression_ID
      - Invalid_expression_ID
      - Invalid_ID_format

  - Доступные тесты для пакета Calculation
    - TestShuntingYard
      - Valid_TwoSum
      - Valid_Expression_with_Priority
      - Valid_Expression_with_Minus
      - Valid_Expression_with_Parentheses
      - Invalid_Expression_(Mismatched_Parentheses)
      - Invalid_Expression_(Unknown_Character)

  - Запуск отдельных тестов:

    ```bash
    go test -v -run=TestCalculateRoute/Valid_expression
    ```

    ```bash
    go test -v -run=TestShuntingYard/Valid_TwoSum
    ```

# Как это работает

Архитекутрно проект состоит из двух частей:

- Сервер (Оркестратор)
- Агент

## Сервер (Оркестратор)

У сервера есть несколько публичных эндпоинтов

- `/api/v1/calculate`
- `/api/v1/expressions`
- `/api/v1/expressions/{id}`
- `/api/v1/register`
- `/api/v1/login`

---

### Принцип работы `/api/v1/calculate`

1) Сервер принимает POST запрос;
2) Декодирует тело из JSON в структуру Request;
3) Делегирует работу над выражением методу handleCalculateRequest;
4) Используя алгоритм [Shunting Yard](https://youtu.be/y_snKkv0gWc?si=Ymv6muB49Du8upEK) он разбивает выражение;
5) После того как выражение было разбито, он формирует задачи, и при необходимости в аргументы подставляет ссылки на зависимые задачи в формате `task{id}` (Именно поэтому у меня arg1 и arg2 строки а не числа);
6) После чего он формирует выражение и добавляет его в базу данных;
7) В конце в слайс добавляются все таски.
![CalcHandler](https://github.com/user-attachments/assets/57b88336-372b-4324-912e-c9c9ffed693d)

### Принцип работы `/api/v1/expressions`

1) Сервер принимает GET запрос;
2) Структуру ExpressionsResponse и добавляет в поле Expressions (слайс из структуры Expression) все выражения для конкретного пользователя;
3) Сортирует выражения по ID;
4) Возвращает JSON массив

### Принцип работы `/api/v1/expressions/{id}`

1) Сервер принимает GET запрос;
2) Проверяет есть ли выражение под таким ID;
3) В зависимости от результата проверки возвращает ошибку или информаицю в JSON в формате.

### Принцип работы агента и сервера

 Агент и сервер соединены между собой протоколом gRPC

1) Сервер смотрит, есть ли у него задачи для агента. При этом он ищёт задачи где два аргумента являются числами, а не ссылками на результат других задач;
2) Сервер отправляет задачу в формате:

```json
{
  "id" id задачи:,
  "expression_id": id выражения,
  "arg1": аргумент 1,
  "arg2": аргумент 2,
  "operation": операция,
  "operation_time": время выполнения,
  "status": статус задачи
}
```

- **POST**

1) Сервер декордирует результат и обновляет задачу в своеё мапе;
2) После чего он ищёт все зависимые от этой задачи другие задачи, и при их наличии заменяет на результат своих вычислений;
3) Если эт опоследняя задача для данного выражения, то он присваивает выражению результат и статус ``done``, после чего удаляет все связанные с этим выражением задачи для освобождения места.

![TaskHandler](https://github.com/user-attachments/assets/099e42f9-d858-44d7-98fc-ebb77dcfa4ec)
![handleTaskget](https://github.com/user-attachments/assets/ade9ba89-d3cc-4830-a6c7-00791df67b13)
![handleTaskPost](https://github.com/user-attachments/assets/233a02ef-bbc0-44a0-9a90-f8d2e7b4702c)

## Агент

### Получение задачи (fetchTask)

1) Агент отправляет gRPC запрос на сервер
2) Если запрос успешен, то задача передается агенту
3) Если запрос неудачен (например, оркестратор недоступен или вернул ошибку), агент логирует ошибку.

### Выполнение задачи (executeTask)

1) Агент выполняет арифметическую операцию, указанную в задаче.
2) Для симуляции времени выполнения задачи используется time.Timer, который ожидает указанное в задаче время (OperationTime).
3) Если операция не может быть выполнена (например, деление на ноль), агент возвращает ошибку.

- Ошибка

  ```json
  {
    "id": id задачи, 
    "result": результат выражения,
    "error": ошибка,
  }
  ```

- Результат

   ```json
  {
    "id": id задачи, 
    "result": результат выражения,
  }
  ```

### Отправка результата (sendResult)

1) Агент отправляет результат выполнения задачи обратно в оркестратор по адресу.
2) Если при выполнении задачи возникла ошибка, она также отправляется в оркестратор.

### Запуск воркеров (RunGoroutines)

1) Агент запускает несколько горутин (воркеров), которые параллельно получают и выполняют задачи.
2) Каждый воркер работает в бесконечном цикле, периодически запрашивая задачи у оркестратора.

# Front-end

К сожалению или счастью я не очень дружу с фронтом, я чисто backend dev 😢
