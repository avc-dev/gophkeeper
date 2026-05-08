package service

// ZeroKey обнуляет ключ в памяти после использования.
func ZeroKey(key []byte) {
	for i := range key {
		key[i] = 0
	}
}
