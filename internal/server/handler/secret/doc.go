// Package secret реализует gRPC-обработчики сервиса секретов GophKeeper.
//
// Обработчики (CreateSecret, GetSecret, ListSecrets, UpdateSecret, DeleteSecret, Ping)
// извлекают userID из контекста (установленного AuthInterceptor),
// делегируют выполнение сервисному слою и конвертируют доменные типы в proto-ответы.
//
// Оптимистичная блокировка реализована через поле expected_version в UpdateSecret:
// при конфликте версий сервис возвращает [domain.ErrVersionConflict],
// который транслируется в gRPC-статус ABORTED.
package secret
