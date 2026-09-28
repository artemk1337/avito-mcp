# Avito MCP

Сервер для Avito Business API на Go 1.27.1. Подключается к MCP-клиенту через stdio. Через него можно читать профиль и объявления, менять цену, читать чаты и отправлять сообщения.

## Установка

Нужны Go 1.27.1 и ключи Avito API. В кабинете продавца откройте «Настройки», затем «Avito API» и «Регистрация нового приложения». Для чатов нужна подходящая подписка и ключ основного аккаунта.

```sh
go build -o avito-mcp ./cmd/avito-mcp
export AVITO_CLIENT_ID=...
export AVITO_CLIENT_SECRET=...
./avito-mcp
```

Если у вас уже есть токен, задайте `AVITO_ACCESS_TOKEN`. Иначе сервер получит и обновит его по `AVITO_CLIENT_ID` и `AVITO_CLIENT_SECRET` через OAuth2 Client Credentials. Если заданы оба варианта, используется готовый токен. Ключи передаются через окружение и не хранятся в репозитории.

Пример настройки MCP-клиента:

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

Для инструментов с `user_id` сначала вызовите `get_profile`. Для создания объявлений Avito использует Автозагрузку.

## Ограничения

Методы взяты из [копии OpenAPI каталога Avito](https://github.com/mkrvas/avito-business-api-reference/blob/master/references/avito-api-openapi.json). [Официальный каталог](https://developers.avito.ru/api-catalog) во время разработки отвечал HTTP 429. [OpenAPI 3.0](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.0.md) описывает формат документации. Реальные запросы к Avito без ключей проверить не удалось.

Avito разрешает до 25 вызовов `list_items` в минуту. Метод не возвращает объявления сотрудников. Для доступа к чужому аккаунту нужна авторизация через OAuth с нужными правами; сервер принимает ключи своего аккаунта или готовый токен.

## Разработка

```sh
go test ./...
```
