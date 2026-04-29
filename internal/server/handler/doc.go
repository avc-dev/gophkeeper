// Package handler содержит gRPC middleware для сервера GophKeeper.
//
// [AuthInterceptor] — unary interceptor, проверяющий JWT-токен из метаданных
// входящего запроса. При успехе добавляет userID в контекст вызова
// через [WithUserID]; обработчики извлекают его через [UserIDFromContext].
//
// Публичные эндпоинты (Register, Login) пропускаются без проверки токена.
package handler
