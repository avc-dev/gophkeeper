# Архитектура GophKeeper

## Обзор

GophKeeper — клиент-серверный менеджер паролей. Клиент работает офлайн с локальной
базой данных, синхронизируя данные с сервером при наличии соединения.

## Стек

| Компонент       | Решение                          | Обоснование                                              |
|-----------------|----------------------------------|----------------------------------------------------------|
| Транспорт       | gRPC + protobuf                  | Бинарный протокол, типизированный API, кодогенерация     |
| Аутентификация  | JWT (EdDSA/Ed25519) + bcrypt     | Stateless токены; Ed25519 — асимметричная подпись вместо HMAC |
| KDF             | Argon2id                         | Победитель Password Hashing Competition, рекомендован OWASP |
| Шифрование      | AES-256-GCM                      | Authenticated encryption: конфиденциальность + целостность |
| БД сервер       | PostgreSQL + pgx                 | Нативный Go-драйвер, без database/sql оверхеда           |
| БД клиент       | SQLite (modernc.org/sqlite)      | Pure Go, без CGO — кросс-компиляция из коробки           |
| Миграции        | migrate                          | Простой инструмент, SQL-файлы без магии                  |
| CLI             | cobra                            | Стандарт де-факто для Go CLI                             |
| Логи            | slog                             | Стандартная библиотека Go 1.21+, без внешних зависимостей |

## Структура проекта

```
gophkeeper/
├── cmd/
│   ├── server/main.go           — точка входа сервера
│   └── client/main.go           — точка входа клиента
├── internal/
│   ├── domain/                  — доменные типы (Secret, User, ошибки)
│   ├── crypto/                  — шифрование AES-256-GCM и KDF Argon2id
│   ├── protoconv/               — конвертеры proto↔domain (SecretType)
│   ├── server/
│   │   ├── app/                 — инициализация сервера, graceful shutdown
│   │   ├── config/              — конфигурация через переменные окружения
│   │   ├── handler/
│   │   │   ├── auth/            — gRPC хендлеры регистрации и входа
│   │   │   ├── secret/          — gRPC хендлеры CRUD секретов
│   │   │   └── middleware.go    — JWT AuthInterceptor
│   │   ├── service/
│   │   │   ├── auth/            — регистрация, логин, валидация токена
│   │   │   └── secret/          — CRUD секретов (тонкий слой над storage)
│   │   └── storage/
│   │       ├── user/            — PostgreSQL: пользователи
│   │       └── secret/          — PostgreSQL: секреты, оптимистичная блокировка
│   └── client/
│       ├── command/
│       │   ├── auth/            — cobra команды register, login, logout
│       │   ├── secret/          — cobra команды add, get, list, delete, copy, sync
│       │   └── cmdutil/         — общие утилиты команд (App, ResolveMasterKey, ...)
│       ├── service/
│       │   ├── auth/            — регистрация, логин, токен, KDF, last_sync_at
│       │   ├── secret/          — CRUD секретов, push/pull, валидация карты
│       │   ├── sync.go          — SyncService (адаптивный фоновый polling)
│       │   ├── key.go           — ZeroKey (обнуление ключа)
│       │   └── grpcmeta.go      — ContextWithBearerToken
│       └── storage/             — SQLite: секреты и авторизационные данные
├── proto/                       — .proto файлы (описание API)
├── migrations/                  — SQL миграции PostgreSQL
└── docs/                        — документация
```

## Архитектурные решения

### Трёхслойная архитектура

Каждая сторона (сервер и клиент) организована в три слоя:

```
Handler/Command → Service → Storage
```

- **Handler/Command** — принимает запрос, валидирует, вызывает Service
- **Service** — бизнес-логика, не знает о транспорте и хранилище
- **Storage** — SQL запросы, не знает о бизнес-логике

Интерфейсы определены только для Storage — это позволяет подменять
реальную БД на in-memory реализацию в тестах. Service намеренно
реализован как конкретная структура без интерфейса.

### Opaque API на границе proto/domain

gRPC хендлеры конвертируют protobuf-структуры в доменные типы сразу
на входе и обратно на выходе. Весь внутренний код работает с доменными
типами, не зависит от protobuf.

### Намеренные упрощения MVP

Данная архитектура намеренно упрощена для скорейшей реализации. Подробнее
о том, что стоит доработать в production — см. [future.md](future.md).

## CLI команды

```
gophkeeper register --email --password
gophkeeper login    --email --password
gophkeeper logout

gophkeeper add credential --name --login --password [--url] [--note]
gophkeeper add card       --name --number --holder --expiry --cvv [--bank] [--note]
gophkeeper add text       --name --content [--note]
gophkeeper add binary     --name --file [--note]

gophkeeper get credential <name>
gophkeeper get card       <name>
gophkeeper get text       <name>
gophkeeper get binary     <name> [--output]

gophkeeper list  [--type credential|card|text|binary]
gophkeeper copy  <type> <name> [--field login|password|number|cvv|content]
gophkeeper delete <type> <name>

gophkeeper sync     # ручная синхронизация
gophkeeper version
```

## Сборка

```bash
make build        # текущая платформа
make build-all    # все платформы (linux/darwin/windows, amd64/arm64)
```

Версия и дата сборки внедряются через ldflags при компиляции.
