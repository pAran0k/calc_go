# calc_go
## Описание
API для вычисления арифметических выражений. Выражение может содержать знаки операций +-/* и ().
Принимает арифметическое выражение через POST-запрос на эндпоинт `/api/v1/calculate` и возвращает результат вычисления либо ошибку.
## Как использовать
### Команда запуска:
`go run ./cmd/main.go`

Сервер запускается на порту `http://localhost:8080`

Отправить запрос можно, используя Postman:

**Запрос**
```json
  {
      "expression" : "2+2"
  }
```
**Ответ**
```json
  {
      "result": 4
  }
```
Либо же, используя curl

**Запрос**
```
  curl --location 'http://localhost:8080/api/v1/calculate' \
  --header 'Content-Type: application/json' \
  --data '{
      "expression" : "2+2"
  }'
```
**Ответ**
```json
  {
      "result": 4
  }
```

## HTTP
| Статус | Код | Метод | Запрос | Ответ |
| ---| --- | --- | --- | --- |
| OK | 200 | POST | { "expression" : "2+2" }| {"result":4} |
| Unprocessable Entity | 422 | POST | { "expression" : "2+2+" } | {"error":"expression is not valid"} |
| Unprocessable Entity | 422 | POST | { "expression" : "2/0" } | {"error":"division by zero"} |
| Unprocessable Entity | 422 | POST | { "expression" : "" } | {"error":"expression is empty"} |
| Bad Request | 400 | POST | {"expression" : "} | {"error":"invalid character '\\r' in string literal"} |
| Method Not Allowed | 405 | GET или другой метод | { "expression" : "2+2" } | not POST method |

## Запуск тестов
`go test pkg/calc_test.go -v `

`go test application/application_test.go -v`

