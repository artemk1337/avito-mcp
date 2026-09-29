# Avito MCP

Сервер для Avito Business API на Go 1.27.1. Подключается к MCP-клиенту через stdio. Через него можно читать профиль и объявления, менять цену, читать чаты и отправлять сообщения.

## Установка

Нужны Go 1.27.1 и ключи Avito API. В кабинете продавца откройте «Настройки», затем «Avito API» и «Регистрация нового приложения». Для чатов нужна подходящая подписка и ключ основного аккаунта.

```sh
go build -o avito-mcp ./cmd/avito-mcp
export AVITO_CLIENT_ID=...
export AVITO_CLIENT_SECRET=...
export AVITO_PROFILE_ID=...
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

`AVITO_PROFILE_ID` - ID аккаунта для аргумента `user_id` у инструментов чатов. Сервер не подставляет его автоматически; MCP-клиент передаёт значение при вызове инструмента. Для чтения переписки `get_profile` не требуется.

### Codex Desktop на macOS

Если сам процесс Codex получает переменные окружения, передайте их серверу по именам через `~/.codex/config.toml`:

```toml
[mcp_servers.avito]
command = "/absolute/path/to/avito-mcp"
env_vars = ["AVITO_CLIENT_ID", "AVITO_CLIENT_SECRET", "AVITO_PROFILE_ID"]
```

В `env_vars` записываются только имена, не значения. Codex, открытый из Dock или Finder, обычно не получает переменные из `~/.zshrc`. Проверьте окружение самого процесса Codex и переподключите MCP после изменения настройки. Наличие переменных в терминале само по себе не подтверждает, что их видит Codex Desktop.

Если передать окружение процессу Codex нельзя, используйте локальный файл с правами `0600`:

```sh
mkdir -p ~/.config/avito-mcp
chmod 700 ~/.config/avito-mcp
cat > ~/.config/avito-mcp/env <<'EOF'
AVITO_CLIENT_ID='your-client-id'
AVITO_CLIENT_SECRET='your-client-secret'
AVITO_PROFILE_ID='your-profile-id'
EOF
chmod 600 ~/.config/avito-mcp/env
```

Настройте `~/.codex/config.toml`, подставив свои абсолютные пути к скрипту и серверу:

```toml
[mcp_servers.avito]
command = "/absolute/path/to/avito-mcp/scripts/run-with-env-file.sh"
args = ["/absolute/path/to/avito-mcp"]
```

Скрипт экспортирует переменные процессу MCP. Не добавляйте файл с ключами в Git. После изменения конфигурации переподключите MCP.

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

Для инструментов с `user_id` передайте `AVITO_PROFILE_ID` в аргумент вызова. `get_profile` доступен как отдельный инструмент, но для работы с чатами не обязателен. Для создания объявлений Авито использует Автозагрузку.

## Ограничения

Методы взяты из [копии OpenAPI каталога Avito](https://github.com/mkrvas/avito-business-api-reference/blob/master/references/avito-api-openapi.json). [OpenAPI 3.0](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.0.md) описывает формат документации. Доступность методов зависит от прав аккаунта и подписки Авито.

Avito разрешает до 25 вызовов `list_items` в минуту. Метод не возвращает объявления сотрудников. Для доступа к чужому аккаунту нужна авторизация через OAuth с нужными правами; сервер принимает ключи своего аккаунта или готовый токен.

## Разработка

```sh
go test ./...
```
