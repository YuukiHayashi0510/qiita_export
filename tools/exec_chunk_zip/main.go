package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	defaultChunkSize = 100
)

func main() {
	targetDir := flag.String("dir", "output", "default value is 'output'")
	targetZip := flag.Int("zip", 0, "specify zip number to process (0 for all)")
	chunkSize := flag.Int("chunk", defaultChunkSize, "number of directories per zip file")
	excludeIDs := flag.String("exclude", "", "comma-separated list of IDs to exclude")
	flag.Parse()

	// ディレクトリ内のエントリを取得
	entries, err := os.ReadDir(*targetDir)
	if err != nil {
		log.Fatal(err)
	}

	// エントリからディレクトリのみ抽出
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(*targetDir, entry.Name()))
		}
	}

	// どのディレクトリのZipかわかりやすくする
	splits := strings.Split(*targetDir, "/")
	baseName := splits[len(splits)-1]

	filteredDirs := dirs
	// excludeIDs が空文字でない場合のみフィルタリングを適用
	if *excludeIDs != "" {
		filteredDirs = filterDirs(dirs, strings.Split(*excludeIDs, ",")...)
		fmt.Printf("Filtered %d directories to %d\n", len(dirs), len(filteredDirs))
	}

	// チャンク数の計算
	numChunks := (len(filteredDirs) + *chunkSize - 1) / *chunkSize

	// チャンク毎にZip化
	for i := range numChunks {
		// 指定されたZIP番号と一致しない場合はスキップ
		if *targetZip != 0 && *targetZip != i+1 {
			continue
		}

		start := i * *chunkSize
		end := min(start+*chunkSize, len(filteredDirs))

		// ZIPファイル名
		zipName := fmt.Sprintf("%s_%d.zip", baseName, i+1)

		// ZIP化 - このチャンクのディレクトリだけを追加
		args := append([]string{"-r", zipName}, filteredDirs[start:end]...)

		cmd := exec.Command("zip", args...)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error creating zip file %s: %v\n", zipName, err)
			continue
		}

		fmt.Printf("Created %s (directories %d-%d of %d)\n", zipName, start+1, end, len(filteredDirs))
	}
}

func filterDirs(dirs []string, excludeIDs ...string) []string {
	filtered := make([]string, 0, len(dirs)-len(excludeIDs))

	for _, v := range dirs {
		isContainedExcludeID := false
		for _, eid := range excludeIDs {
			if strings.Contains(v, eid) {
				isContainedExcludeID = true
				break
			}
		}

		if !isContainedExcludeID {
			fmt.Println(filepath.Base(v))
			filtered = append(filtered, v)
		}
	}

	return filtered
}
