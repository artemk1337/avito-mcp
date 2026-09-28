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

Без ключей сервер запускается и показывает инструменты, но запросы к Avito возвращают ошибку до настройки авторизации.

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

### Codex Desktop на macOS

Приложение, открытое из Dock или Finder, может не получить переменные из `~/.zshrc`. Для него храните ключи в локальном файле `~/.config/avito-mcp/env` с правами `0600`:

```sh
mkdir -p ~/.config/avito-mcp
chmod 700 ~/.config/avito-mcp
cat > ~/.config/avito-mcp/env <<'EOF'
AVITO_CLIENT_ID='your-client-id'
AVITO_CLIENT_SECRET='your-client-secret'
EOF
chmod 600 ~/.config/avito-mcp/env
```

Настройте `~/.codex/config.toml`, подставив свои абсолютные пути к скрипту и серверу:

```toml
[mcp_servers.avito]
command = "/absolute/path/to/avito-mcp/scripts/run-with-env-file.sh"
args = ["/absolute/path/to/avito-mcp"]
```

Скрипт экспортирует ключи только процессу MCP. Не добавляйте файл с ключами в Git. Если Codex уже запустил MCP-сервер без ключей, переподключите его после изменения конфигурации.

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

Методы взяты из [копии OpenAPI каталога Avito](https://github.com/mkrvas/avito-business-api-reference/blob/master/references/avito-api-openapi.json). [OpenAPI 3.0](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.0.md) описывает формат документации. Доступность методов зависит от прав аккаунта и подписки Авито.

Avito разрешает до 25 вызовов `list_items` в минуту. Метод не возвращает объявления сотрудников. Для доступа к чужому аккаунту нужна авторизация через OAuth с нужными правами; сервер принимает ключи своего аккаунта или готовый токен.

## Разработка

```sh
go test ./...
```
