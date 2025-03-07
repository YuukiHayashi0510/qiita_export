package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 探索するファイルパスをflagで受け取る
	var rootPath string
	flag.StringVar(&rootPath, "dir", "", "探索するディレクトリパス")
	flag.Parse()

	if rootPath == "" {
		fmt.Println("ディレクトリ指定は必須です")
		os.Exit(1)
	}

	// ファイル探索
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// ファイル名（拡張子なし）とディレクトリを取得
		baseName := strings.TrimSuffix(filepath.Base(path), ".md")
		dirPath := filepath.Dir(path)

		// 対応するメタデータファイルのパスを取得
		metadataPath := filepath.Join(dirPath, baseName+"_metadata.json")

		// メタデータファイルが存在するか確認
		if _, err := os.Stat(metadataPath); err != nil {
			return fmt.Errorf("メタデータファイルがありません: %s", metadataPath)
		}

		// メタデータファイルへのリンクをマークダウンに追加
		return appendMetadataLinkToMarkdown(path, baseName+"_metadata.json")
	})

	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}
}

// マークダウンファイルにメタデータファイルへのリンクを追加する関数
func appendMetadataLinkToMarkdown(mdPath string, metadataFileName string) error {
	// マークダウンファイルを読み込む
	mdContent, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("マークダウンファイル読み込みエラー: %w", err)
	}

	// メタデータリンクがすでに付与されている場合はスキップ
	if strings.Contains(string(mdContent), "["+metadataFileName+"]") {
		return nil
	}

	// メタデータファイルへのリンクを付与したコンテンツを作成
	updatedContent := string(mdContent) +
		"\n[" + metadataFileName + "](" + metadataFileName + ")\n"

	// ファイルに書き戻す
	if err = os.WriteFile(mdPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("ファイル書き込みエラー: %w", err)
	}

	fmt.Printf("メタデータリンク追加完了 file=%s\n", filepath.Base(mdPath))
	return nil
}
