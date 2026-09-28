# Avito MCP

MCP-сервер на Go 1.27.1 для Avito Business API. Работает через stdio и предоставляет инструменты для профиля, объявлений и чатов.

## Установка

Нужны Go 1.27.1 и ключи Avito API. Ключи можно создать в кабинете продавца: **Настройки → Avito API → Регистрация нового приложения**. Для Messenger API нужна подходящая подписка и ключ основного аккаунта.

```sh
go build -o avito-mcp ./cmd/avito-mcp
export AVITO_CLIENT_ID=...
export AVITO_CLIENT_SECRET=...
./avito-mcp
```

Сервер также принимает готовый `AVITO_ACCESS_TOKEN` вместо пары `AVITO_CLIENT_ID` / `AVITO_CLIENT_SECRET`. Токен имеет приоритет. При использовании пары ключей сервер сам получает и обновляет токен по OAuth2 Client Credentials. Секреты не записываются в репозиторий.

Пример конфигурации MCP-клиента:

```json
{
  "mcpServers": {
    "avito": {
      "command": "/absolute/path/to/avito-mcp",
      "env": {
        "AVITO_CLIENT_ID": "your-client-id",
        "AVITO_CLIENT_SECRET": "your-client-secret"
      }
    }
  }
}
```

## Инструменты

| Инструмент | Действие |
| --- | --- |
| `get_profile` | Получить профиль и ID аккаунта |
| `list_items` | Найти свои объявления с фильтрами и пагинацией |
| `get_item` | Получить объявление по ID |
| `update_item_price` | Изменить цену объявления |
| `list_chats` | Получить список чатов |
| `get_chat` | Получить чат |
| `list_messages` | Получить сообщения без пометки чата прочитанным |
| `send_message` | Отправить текстовое сообщение |
| `mark_chat_read` | Пометить чат прочитанным |

Для инструментов с `user_id` сначала вызовите `get_profile`. Создание объявлений через REST API здесь не предусмотрено: Avito использует для этого Автозагрузку.

## Ограничения

Методы сверены по [копии OpenAPI каталога Avito](https://github.com/mkrvas/avito-business-api-reference/blob/master/references/avito-api-openapi.json), так как [официальный каталог](https://developers.avito.ru/api-catalog) из среды разработки возвращал HTTP 429. Ссылка на [OpenAPI 3.0](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.0.md) описывает формат спецификации, а не API Avito. Реальные запросы к Avito требуют ключей и не проверялись.

`list_items` ограничен Avito до 25 запросов в минуту и не возвращает объявления сотрудников. Для доступа к другим аккаунтам нужен отдельный OAuth Authorization Code flow и соответствующие scopes; этот сервер работает с ключом собственного аккаунта или заранее полученным токеном.

## Разработка

```sh
go test ./...
```
