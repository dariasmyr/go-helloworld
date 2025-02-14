package sort

import (
	"fmt"
	"sort"
	"testing"
)

func quickSort(arr []string) []string {
	if len(arr) <= 1 {
		return arr
	}
	fmt.Println("Initial array", arr)

	pivot := arr[len(arr)-1]
	fmt.Println("Pivot element", pivot)
	var left []string
	var right []string

	for _, value := range arr[:len(arr)-1] {
		if value <= pivot {
			left = append(left, value)
		} else {
			right = append(right, value)
		}
	}

	fmt.Println("Left", left)
	fmt.Println("Right", right)

	fmt.Println("Recursive call for left", left)
	sortedLeft := quickSort(left)
	fmt.Println("Sorted left", sortedLeft)

	fmt.Println("Recursive call for right", right)
	sortedRight := quickSort(right)
	fmt.Println("Sorted right", sortedRight)

	return append(append(sortedLeft, pivot), sortedRight...)
}

func topKFrequent(words []string, k int) []string {

	sortedArray := quickSort(words)

	fmt.Printf("SortedArray: %v\n", sortedArray)

	wordsMap := map[string]int{}

	for _, word := range sortedArray {
		wordsMap[word]++
	}

	fmt.Printf("WordsMap: %v", wordsMap)

	type wordFreq struct {
		word  string
		count int
	}

	var wordList []wordFreq

	for word, freq := range wordsMap {
		wordList = append(wordList, wordFreq{word: word, count: freq})
	}

	sort.Slice(wordList, func(i, j int) bool {
		return wordList[i].count > wordList[j].count
	})

	result := make([]string, 0, k)
	for i := 0; i < k; i++ {
		result = append(result, wordList[i].word)
	}

	return result
}

func quickSortStr(words []string) []string {
	if len(words) == 0 {
		return words
	}

	pivot := words[len(words)-1]

	var left []string
	var right []string

	for i := 0; i < len(words)-1; i++ {
		if words[i] >= pivot {
			right = append(right, words[i])
		} else {
			left = append(left, words[i])
		}
	}

	sortedLeft := quickSort(left)
	sortedRight := quickSort(right)

	return append(append(sortedLeft, pivot), sortedRight...)
}

func testSort() {
	arr := []string{"o", "a", "b", "g", "k"}
	fmt.Println("Arr", arr)

	sortedArr := quickSort(arr)
	fmt.Println("sortedArr", sortedArr)
}

func TestQuickSort(t *testing.T) {
	testSort()
}
