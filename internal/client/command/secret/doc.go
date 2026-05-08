// Package secret содержит cobra-команды для работы с секретами клиента GophKeeper.
//
// Команды: add (credential/card/text/binary), get (credential/card/text/binary),
// list, delete, copy (в буфер обмена), sync (ручная синхронизация с сервером).
// Все команды, требующие расшифровки, принимают --master-password
// и запрашивают его интерактивно при отсутствии.
package secret
