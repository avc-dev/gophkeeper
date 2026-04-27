package command

import "time"

// nowUTC возвращает текущее время в UTC (вынесено для удобства тестирования).
func nowUTC() time.Time { return time.Now().UTC() }
