// Package config загружает конфигурацию gRPC-сервера GophKeeper из переменных окружения.
//
// Переменные окружения:
//   - GRPC_ADDR — адрес для прослушивания (по умолчанию :8080).
//   - DATABASE_DSN — строка подключения к PostgreSQL.
//   - JWT_PRIVATE_KEY_PATH — путь к файлу Ed25519 приватного ключа (PEM).
//   - JWT_PUBLIC_KEY_PATH  — путь к файлу Ed25519 публичного ключа (PEM).
package config
