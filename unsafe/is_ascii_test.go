package unsafe

import (
	"testing"
	"unsafe"
)

// isASCIIFast проверяет, содержит ли байтовый срез ТОЛЬКО ASCII-символы (байты с младшими 7 битами).
// Использует unsafe и побитовые операции для ускорения на 64-битных платформах.
func isASCIIFast(s []byte) bool {
	n := len(s)
	i := 0

	// Обрабатываем срез блоками по 8 байт (uint64), пока возможно:
	for ; i+8 <= n; i += 8 {
		// Чтение 8 байт как одного uint64 без копирования (unsafe).
		block := *(*uint64)(unsafe.Pointer(&s[i]))

		// Если в каком-либо из 8 байт старший бит (0x80) установлен — это НЕ ASCII.
		// Маска 0x8080808080808080 проверяет старший бит в каждом байте блока.
		if block&0x8080808080808080 != 0 {
			return false
		}
	}

	// Проверяем оставшиеся байты по одному:
	for ; i < n; i++ {
		// Если старший бит установлен — не ASCII.
		if s[i]&0x80 != 0 {
			return false
		}
	}

	// Все байты имели старший бит = 0 → это ASCII.
	return true
}

// isAllRunesASCII проверяет, состоит ли байтовый срез из символов, которые представлены ТОЛЬКО ASCII‑рунами.
// Реализует пошаговую валидацию UTF-8 и считает любую многобайтовую руну НЕ ASCII.
func isAllRunesASCII(b []byte) bool {
	for i := 0; i < len(b); {
		first := b[i]

		// 1) ASCII-руна: первый бит равен 0 → одиночный байт (стандарт ASCII).
		if first&0x80 == 0 {
			i++
			continue
		}

		// 2) Определяем длину UTF-8-последовательности по начальному байту:
		//    110xxxxx → 2 байта
		//    1110xxxx → 3 байта
		//    11110xxx → 4 байта
		var n int
		switch {
		case first&0xE0 == 0xC0:
			n = 2
		case first&0xF0 == 0xE0:
			n = 3
		case first&0xF8 == 0xF0:
			n = 4
		default:
			// Некорректный первый байт UTF-8 (не попадает ни в один допустимый диапазон).
			return false
		}

		// 3) Проверяем, хватает ли байт в срезе для всей руны.
		if i+n > len(b) {
			return false
		}

		// 4) Проверяем, что оставшиеся байты соответствуют формату 10xxxxxx (UTF-8 continuation bytes).
		for j := 1; j < n; j++ {
			if b[i+j]&0xC0 != 0x80 {
				return false
			}
		}

		// 5) Любая многобайтовая руна ≠ ASCII → сразу false.
		return false
	}

	// Все символы прошли как одиночные ASCII-руны.
	return true
}

func isASCIIPlain(s []byte) bool {
	for _, b := range s {
		if b&0x80 != 0 {
			return false
		}
	}
	return true
}

func isASCIIBytesIndex(s []byte) bool {
	for i := range s {
		if s[i] >= 0x80 { // equal to s[i] >= 128
			return false
		}
	}
	return true
}

func TestIsASCIIFast(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", true},
		{"ASCII only", "Hello, World!", true},
		{"Extended Latin é", "Café", false},
		{"Cyrillic", "Привет", false},
		{"Emoji", "😀", false},
		{"Mixed ASCII and Unicode", "Hello, 世界", false},
		{"All ASCII boundary", string([]byte{0x00, 0x7F}), true},
		{"Non-ASCII boundary", string([]byte{0x00, 0x80}), false},
		{"Long ASCII string", "This is a long ASCII string with 128 characters! ................................................................", true},
		{"Long string with Unicode", "This is a long string with Unicode: Привет мир!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isASCIIFast([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("isASCIIFast(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkIsASCIIFast(b *testing.B) {
	data := make([]byte, 1024*1024) // 1MB ASCII
	for i := 0; i < len(data); i++ {
		data[i] = 'A'
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !isASCIIFast(data) {
			b.Fatal("failed")
		}
	}
}

func BenchmarkIsASCIIBytesIndex(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := 0; i < len(data); i++ {
		data[i] = 'A'
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !isASCIIBytesIndex(data) {
			b.Fatal("failed")
		}
	}
}

func BenchmarkIsASCIIPlain(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := 0; i < len(data); i++ {
		data[i] = 'A'
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !isASCIIPlain(data) {
			b.Fatal("failed")
		}
	}
}

func BenchmarkIsAllRunesASCII(b *testing.B) {
	data := make([]byte, 1024*1024)
	for i := 0; i < len(data); i++ {
		data[i] = 'A'
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !isAllRunesASCII(data) {
			b.Fatal("failed")
		}
	}
}
