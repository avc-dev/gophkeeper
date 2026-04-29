// Package domain содержит доменные типы и sentinel-ошибки приложения GophKeeper.
//
// Пакет не имеет зависимостей от инфраструктурных слоёв (gRPC, SQL, HTTP).
// Все остальные пакеты ссылаются на domain, но не наоборот.
//
// Основные типы:
//   - [Secret] — зашифрованный секрет пользователя с метаданными версионирования.
//   - [SecretType] — перечисление типов: credential, card, text, binary.
//   - [User] — учётная запись пользователя сервера.
//
// Sentinel-ошибки ([ErrSecretNotFound], [ErrVersionConflict], [ErrEmailTaken])
// намеренно экспортируются, чтобы верхние слои могли сравнивать их через errors.Is.
package domain
