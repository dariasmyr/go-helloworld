package csv_merger

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"testing"
)

func TestCSVMerger(t *testing.T) {
	t.Run("Convert XMT to CSV", func(t *testing.T) {
		file1 := "data/TelegramAds_for_cTrader_AD00000015_202503261458_5min.csv"
		file2 := "data/TelegramAds_for_cTrader_AD00000015_budget_202503261459_5min.csv"
		outputFile := "data/merged.csv"

		data1, err := readCSV(file1)
		if err != nil {
			fmt.Println("Error reading file1:", err)
			return
		}

		data2, err := readCSV(file2)
		if err != nil {
			fmt.Println("Error reading file2:", err)
			return
		}

		mergedData := mergeCSV(data1, data2)
		writeCSV(outputFile, mergedData)
	})
}

func readCSV(filename string) (map[string][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	data := make(map[string][]string)
	for i, row := range rows {
		if i == 0 {
			data["header"] = row
			continue
		}
		key := row[0]
		data[key] = row[1:]
	}
	return data, nil
}

func mergeCSV(data1, data2 map[string][]string) [][]string {
	merged := [][]string{}

	header := append(data1["header"], data2["header"][1:]...)
	merged = append(merged, header)

	keys := make(map[string]bool)
	for k := range data1 {
		if k != "header" {
			keys[k] = true
		}
	}
	for k := range data2 {
		if k != "header" {
			keys[k] = true
		}
	}

	sortedKeys := []string{}
	for k := range keys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		row := []string{key}
		if val, ok := data1[key]; ok {
			row = append(row, val...)
		} else {
			row = append(row, make([]string, len(data1["header"])-1)...) // Fill with empty values
		}
		if val, ok := data2[key]; ok {
			row = append(row, val...)
		} else {
			row = append(row, make([]string, len(data2["header"])-1)...) // Fill with empty values
		}
		merged = append(merged, row)
	}

	return merged
}

func writeCSV(filename string, data [][]string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	defer writer.Flush()

	for _, row := range data {
		if err := writer.Write(row); err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}
}
