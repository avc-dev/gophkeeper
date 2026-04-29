// Package protoconv предоставляет конвертеры между доменными типами и protobuf-типами.
//
// Функции пакета используются в handler и client/service слоях для трансляции
// на границе proto↔domain, сохраняя opaque API: вышестоящие слои работают
// только с [domain] типами, а не с pb.* напрямую.
package protoconv
