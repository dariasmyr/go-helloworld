package substring

import (
	"fmt"
	"testing"
)

func findSubstring(s, sub string) int {
	n := len(s)
	m := len(sub)
	result := 0

	if m == 0 || m > n {
		return result
	}

	for i := 0; i <= n-m; {
		fmt.Printf("Processing %s\n", s[i:i+m])
		if s[i:i+m] == sub {
			result++
			i += m
		} else {
			i++
		}
	}

	return result
}

func TestFindSubstring(t *testing.T) {
	substringCount := findSubstring("gafggafggafggafg", "fg")
	fmt.Println("substringCount:", substringCount)
}
