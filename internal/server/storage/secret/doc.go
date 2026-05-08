// Package secret реализует PostgreSQL-хранилище секретов сервера GophKeeper.
//
// Все операции изолированы по userID: пользователь может обращаться только
// к своим секретам. Запросы используют pgx/v5 и pgxpool для эффективного
// управления соединениями.
//
// Оптимистичная блокировка в [Storage.Update] реализована через условие
// WHERE version = expected_version; при несовпадении возвращается [domain.ErrVersionConflict].
//
// Мягкое удаление не применяется: Delete физически удаляет строку;
// клиент синхронизирует список через ListSecrets с фильтром by updated_at (since).
package secret
