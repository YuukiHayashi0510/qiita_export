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
	chunkSize = 100
)

func main() {
	targetDir := flag.String("dir", "output", "default value is 'output'")
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

	// チャンク数の計算
	numChunks := (len(dirs) + chunkSize - 1) / chunkSize

	// チャンク毎にZip化
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(dirs) {
			end = len(dirs)
		}

		// ZIPファイル名
		zipName := fmt.Sprintf("%s_%d.zip", baseName, i+1)

		// ZIP化 - 対象のディレクトリを直接追加
		// コマンドの引数を用意
		args := append([]string{"-r", zipName}, dirs[start:end]...)

		cmd := exec.Command("zip", args...)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error creating zip file %s: %v\n", zipName, err)
			continue
		}

		fmt.Printf("Created %s (directories %d-%d of %d)\n", zipName, start+1, end, len(dirs))
	}
}
