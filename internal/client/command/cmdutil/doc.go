// Package cmdutil предоставляет общие утилиты для cobra-команд клиента GophKeeper.
//
// Содержит тип [App] — контейнер инициализированных сервисов, разделяемый между командами.
// Вспомогательные функции: [App.ResolveMasterKey] (промпт + деривация ключа),
// [App.AuthedContext] (gRPC-контекст с JWT), [ReadPassword] (ввод без эха),
// [AddMasterPasswordFlag] (регистрация общего флага --master-password).
package cmdutil
