# GophKeeper

[![Go](https://img.shields.io/github/go-mod/go-version/avc-dev/gophkeeper?logo=go)](go.mod)
![License](https://img.shields.io/badge/license-MIT-green)
[![Coverage](https://codecov.io/gh/avc-dev/gophkeeper/branch/main/graph/badge.svg)](https://codecov.io/gh/avc-dev/gophkeeper)

Клиент-серверный менеджер паролей с zero-knowledge архитектурой. Все секреты шифруются на клиенте до отправки на сервер — сервер никогда не видит данные в открытом виде.

Клиент работает офлайн с локальной базой данных и синхронизирует данные с сервером при наличии соединения.

## Возможности

- Хранение паролей, банковских карт, текстовых заметок и произвольных файлов
- End-to-end шифрование (AES-256-GCM) с ключом, производным от мастер-пароля
- Автоматическая синхронизация между устройствами через gRPC
- Офлайн-режим с полным доступом к данным
- Кросс-платформенный CLI: macOS, Linux, Windows (amd64/arm64)
- Копирование паролей в буфер обмена без отображения на экране

## Документация

| Документ | Содержание |
|----------|------------|
| [Архитектура](docs/architecture.md) | Стек, структура проекта, CLI команды |
| [Безопасность](docs/security.md) | Иерархия ключей, модель угроз, что где хранится |
| [Синхронизация](docs/sync.md) | Алгоритм PUSH/PULL, разрешение конфликтов |
| [Будущие улучшения](docs/future.md) | Что намеренно упрощено в MVP |

## Быстрый старт

```bash
# Запустить PostgreSQL
docker compose up -d

# Накатить миграции
make migrate-up

# Собрать
make build

# Зарегистрироваться и начать работу
./bin/client register --email user@example.com --password
./bin/client add credential --name github --login user --password secret
./bin/client list
```

## Сборка

```bash
make build        # текущая платформа → bin/
make build-all    # все платформы → dist/
```

## Разработка

**Зависимости:** Go 1.22+, Docker, protoc, golangci-lint, migrate

```bash
# Поднять окружение
docker compose up -d              # PostgreSQL для разработки
make migrate-up                   # применить миграции

# Тесты
make test-unit                    # юнит-тесты, проверка покрытия ≥ 70%
make test-integration             # интеграционные тесты (поднимает отдельный Docker)

# Кодогенерация
make proto                        # сгенерировать Go-код из .proto файлов

# Качество кода
make lint                         # запустить golangci-lint
```

Версия и дата сборки внедряются через ldflags при компиляции:

```bash
./bin/client version
# GophKeeper v1.2.0 (built 2026-04-24T10:00:00Z)
```
