package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Replacement 定義
type Replacement struct {
	OldURL string
	NewURL string
	Count  int
}

func main() {
	// コマンドライン引数を定義
	rootDir := flag.String("dir", "", "Directory to scan for markdown files")
	csvFile := flag.String("csv", "", "CSV file with replacement mappings")
	flag.Parse()

	if *rootDir == "" || *csvFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// CSVファイルを読み込む
	replacements, err := loadReplacements(*csvFile)
	if err != nil {
		fmt.Printf("Error loading replacements: %v\n", err)
		os.Exit(1)
	}

	err = filepath.WalkDir(*rootDir, func(path string, d os.DirEntry, err error) error {
		return processFile(path, d, err, replacements)
	})
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// 結果の表示
	printResults(replacements)
}

func loadReplacements(csvFile string) ([]Replacement, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file: %v", err)
	}

	var replacements []Replacement
	for _, record := range records {
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid record in CSV file: %v", record)
		}
		replacements = append(replacements, Replacement{
			OldURL: record[0],
			NewURL: record[1],
			Count:  0,
		})
	}

	return replacements, nil
}

func processFile(path string, d os.DirEntry, err error, replacements []Replacement) error {
	if err != nil {
		return err
	}

	// ディレクトリまたは.mdファイル以外はスキップ
	if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
		return nil
	}

	// ファイルの内容を読み込み
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", path, err)
	}

	// 文字列置換を実行
	newContent := string(content)
	fileChanged := false

	for i, replacement := range replacements {
		// 置換前の文字列を数える
		count := strings.Count(newContent, replacement.OldURL)
		if count > 0 {
			newContent = strings.ReplaceAll(newContent, replacement.OldURL, replacement.NewURL)
			replacements[i].Count += count
			fileChanged = true
		}
	}

	// 置換が行われた場合のみファイルを更新
	if fileChanged {
		err = os.WriteFile(path, []byte(newContent), d.Type())
		if err != nil {
			return fmt.Errorf("error writing file %s: %v", path, err)
		}
	}

	return nil
}

func printResults(replacements []Replacement) {
	fmt.Println("\n置換結果:")
	totalCount := 0

	for _, replacement := range replacements {
		if replacement.Count > 0 {
			fmt.Printf("%s -> %d件\n", replacement.OldURL, replacement.Count)
			totalCount += replacement.Count
		}
	}

	fmt.Printf("\n合計: %d件\n", totalCount)
}
