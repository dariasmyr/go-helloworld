package substring

import (
	"fmt"
	"testing"
)

func findSubstring(s, sub string) bool {
	n := len(s)
	m := len(sub)

	if m == 0 {
		return true
	}

	if m > n {
		return false
	}

	for i := 0; i <= n-m; i++ {
		fmt.Printf("Processing %s\n", s[i:i+m])
		if s[i:i+m] == sub {
			return true
		}
	}

	return false
}

func TestFindSubstring(t *testing.T) {
	substringFound := findSubstring("ofdgadfgagfafgafgofdgadfgagfafgafgofdgadfgagfafgafg", "gafg")

	fmt.Println("substringFound", substringFound)
}
