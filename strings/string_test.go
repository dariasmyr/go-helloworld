package strings

import (
	"fmt"
	"testing"
)

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}

	isNegative := false
	if n < 0 {
		isNegative = true
		n = -n
	}

	var buf [20]byte
	i := len(buf)

	for n > 0 { // Start from the env bo the duf array
		i--
		buf[i] = byte(n%10) + '0' // We take the remainder of the division by 10 to get the last digit of the number.
		n /= 10                   // Divide by 10 to remove the last digit
	}

	if isNegative {
		i--
		buf[i] = '-'
	}

	return string(buf[i:])
}

func TestStrConv(t *testing.T) {
	fmt.Println(intToStr(12345))
	fmt.Println(intToStr(-9876))
	fmt.Println(intToStr(0))
}

func findUnique(n []int) []int {
	j := 0
	for i := 1; i < len(n); i++ {
		fmt.Printf("Comparing j=%d and i=%d\n", n[j], n[i])
		if n[j] != n[i] {
			j++
			n[j] = n[i]
		}
		fmt.Printf("Updated array n=%v\n", n)
	}

	return n[:j+1]
}

func Test(t *testing.T) {
	slice := []int{1, 1, 2, 2, 3, 4, 4, 5}
	fmt.Println(findUnique(slice))
}
