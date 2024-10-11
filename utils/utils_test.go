package utils_test

import (
	"bufio"
	"fmt"
	fs "gaffeine/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestCeilingPowerOfTwo32(t *testing.T) {
	numbers, expecteds := readLines("test_files/10w_ceilingPowerOfTwo32.txt")
	length := len(numbers)
	for i := 0; i < length; i++ {
		num, expected := numbers[i], expecteds[i]
		relt := fs.CeilingPowerOfTwo32(num)
		assert.Equal(t, expected, relt, fmt.Sprintf("integer.CeilingPowerOfTwo(%d)=%d, 现在=%d", num, expected, relt))
	}
}

func TestCeilingPowerOfTwo64(t *testing.T) {
	numbers, expecteds := readLines("test_files/10w_ceilingPowerOfTwo32.txt")
	length := len(numbers)
	for i := 0; i < length; i++ {
		num, expected := numbers[i], expecteds[i]
		relt := fs.CeilingPowerOfTwo64(int64(num))
		assert.Equal(t, int64(expected), relt, fmt.Sprintf("long.CeilingPowerOfTwo(%d)=%d, 现在=%d", num, expected, relt))
	}
}

func readLines(fname string) ([]int, []int) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, nil
	}

	func(file *os.File) { defer file.Close() }(f)

	numbers, expecteds := make([]int, 0), make([]int, 0)

	for scanner := bufio.NewScanner(f); scanner.Scan(); {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		vs := strings.Split(line, ",")

		if len(vs) != 2 {
			continue
		}

		num, _ := strconv.Atoi(vs[0])
		expected, _ := strconv.Atoi(vs[1])

		numbers = append(numbers, num)
		expecteds = append(expecteds, expected)
	}
	return numbers, expecteds
}
