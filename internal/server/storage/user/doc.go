// Package user реализует PostgreSQL-хранилище пользователей сервера GophKeeper.
//
// Предоставляет две операции: [Storage.Create] (регистрация) и [Storage.FindByEmail] (поиск при входе).
// При попытке создать пользователя с уже занятым email возвращается [domain.ErrEmailTaken].
package user
