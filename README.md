# Распределённый вычислитель арифметических выражений
## Описание
с помощью Оркестратора и Агента параллельно вычисляет арифметические выражения. 

Оркестратор принимает выражение, разбивает его на подзадачи и через http отдает их Агенту.

Агент принимает задачи от Оркестратора, обрабатывает их, и возвращает результат оркестратору.
## Как использовать
### Запуск Оркестратора:
`go run ./cmd/orchestrator/main.go`

### Запуск Агента:
`go run ./cmd/agent/main.go`

Сервер запускается на порту `http://localhost:8080`

## Эндпоинты:
### 1.Оркестратор принимает выражение на эндпоинт:
```bash
 POST /api/v1/calculate
```
Пример запроса:

```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2"
}'
```

Успешный ответ (201):

```json
{
    "id": "1"
}
```

### 2. Получение списка выражений

```bash
GET /api/v1/expressions
```

Пример ответа (200):

```json
{
    "expressions": [
        {
            "name": "2+2",
            "status": 0,
            "id": 1,
            "result": 4,
            "node": {
                "value": "+",
                "left": {
                    "value": "2.000000"
                },
                "right": {
                    "value": "2.000000"
                }
            }
        },
        {
            "name": "2/0",
            "status": 3,
            "id": 2,
            "result": 0,
            "node": {
                "value": "/",
                "left": {
                    "value": "2.000000"
                },
                "right": {
                    "value": "0.000000"
                }
            }
        }
    ]
}
```

### 3. Получение выражения по ID

```bash
GET /api/v1/expressions/{id}
```

Пример запроса:

```bash
curl http://localhost:8080/api/v1/expressions/1
```

Ответ (200):

```json
{
    "expression": {
        "name": "2+2",
        "status": 0,
        "id": 1,
        "result": 4,
        "node": {
            "value": "+",
            "left": {
                "value": "2.000000"
            },
            "right": {
                "value": "2.000000"
            }
        }
    }
}
```



# Переменные окружения
- `TIME_ADDITION_MS` - время сложения (мс)
- `TIME_SUBTRACTION_MS` - время вычитания (мс)
- `TIME_MULTIPLICATIONS_MS` - время умножения (мс)
- `TIME_DIVISIONS_MS` - время деления (мс)
- `ORCHESTRATOR_ADDR` - URL оркестратора
- `COMPUTING_POWER` - количество параллельных задач


## Запуск тестов
`go test internal\services\agent\agent_test.go -v `


