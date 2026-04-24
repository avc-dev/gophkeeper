# Архитектура GophKeeper

## Обзор

GophKeeper — клиент-серверный менеджер паролей. Клиент работает офлайн с локальной
базой данных, синхронизируя данные с сервером при наличии соединения.

## Стек

| Компонент       | Решение                          | Обоснование                                              |
|-----------------|----------------------------------|----------------------------------------------------------|
| Транспорт       | gRPC + protobuf                  | Бинарный протокол, типизированный API, кодогенерация     |
| Аутентификация  | JWT + bcrypt                     | Stateless токены, не требуют хранения сессий на сервере  |
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
│   ├── server/main.go       — точка входа сервера
│   └── client/main.go       — точка входа клиента
├── internal/
│   ├── domain/              — доменные типы (Secret, User, Card...)
│   ├── crypto/              — шифрование и KDF
│   ├── server/
│   │   ├── handler/         — gRPC хендлеры, валидация запросов
│   │   ├── service/         — бизнес-логика сервера
│   │   └── storage/         — SQL запросы к PostgreSQL
│   └── client/
│       ├── command/         — cobra команды
│       ├── service/         — бизнес-логика клиента и синхронизация
│       └── storage/         — SQL запросы к SQLite
├── proto/                   — .proto файлы (описание API)
├── migrations/              — SQL миграции PostgreSQL
└── docs/                    — документация
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
gophkeeper login --email --password
gophkeeper add credential --name --login --password [--url] [--meta]
gophkeeper add card --name --number --holder --expiry --cvv [--bank] [--meta]
gophkeeper add text --name [--content|--file] [--meta]
gophkeeper add binary --name --file [--meta]
gophkeeper get <type> <name> [--show-cvv] [--output]
gophkeeper copy <type> <name> [--field]
gophkeeper list [--type]
gophkeeper search <query>
gophkeeper edit <type> <name> [флаги]
gophkeeper delete <type> <name>
gophkeeper version
```

## Сборка

```bash
make build        # текущая платформа
make build-all    # все платформы (linux/darwin/windows, amd64/arm64)
```

Версия и дата сборки внедряются через ldflags при компиляции.
