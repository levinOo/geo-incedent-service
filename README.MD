# geo-incident-service

Сервис (WIP) для проверки координат пользователя относительно опасных зон (инцидентов) и асинхронной отправки вебхуков на внешний портал. [conversation_history:1]

## Возможности (план)
- CRUD для инцидентов (роль оператора, API-key). [conversation_history:1]
- Проверка координат: вернуть ближайшие/пересекающиеся опасные зоны + логирование проверки. [conversation_history:1]
- Асинхронная отправка вебхуков через очередь (Redis) + retry при ошибках доставки. [conversation_history:1]
- Статистика: уникальные `user_id` за последние N минут. [conversation_history:1]
- Health-check для мониторинга. [conversation_history:1]

## Стек
- Go 1.24+ (Clean Architecture: Handler → Service/Usecase → Repository). [conversation_history:1]
- PostgreSQL 15 (геоданные через PostGIS — решение для удобной гео-логики). [conversation_history:1]
- Redis (очередь вебхуков + кэш/статистика). [conversation_history:1]
- Docker / docker-compose. [conversation_history:1]

## Быстрый старт (когда будет готово)
```bash
cp .env.example .env
docker compose up --build
