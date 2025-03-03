package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// メタデータJSONの構造体
type Metadata struct {
	ID string `json:"id"`
}

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

		// メタデータからIDを取得
		id, err := getIDFromMetadata(metadataPath)
		if err != nil {
			return err
		}

		// IDをマークダウンに追加
		return appendIDtoMarkdown(path, id)
	})

	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}
}

// マークダウンファイルにIDを追加する関数
// getIDFromMetadataの呼び出しを削除し、idを引数として受け取るように変更
func appendIDtoMarkdown(mdPath string, id string) error {
	// マークダウンファイルを読み込む
	mdContent, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("マークダウンファイル読み込みエラー: %w", err)
	}

	// IDがすでに付与されている場合はスキップ
	if strings.Contains(string(mdContent), "ID: "+id) {
		return nil // 静かにスキップ
	}

	// IDを付与したコンテンツを作成
	updatedContent := string(mdContent) + "\nID: " + id + "\n"

	// ファイルに書き戻す
	if err = os.WriteFile(mdPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("ファイル書き込みエラー: %w", err)
	}

	fmt.Printf("ID追加完了 title=%s, id=%s\n", filepath.Base(mdPath), id)
	return nil
}

// メタデータファイルからIDを取得する関数
func getIDFromMetadata(metadataPath string) (string, error) {
	// メタデータファイルが存在するか確認
	if _, err := os.Stat(metadataPath); err != nil {
		return "", fmt.Errorf("メタデータファイルがありません: %s", metadataPath)
	}

	// メタデータファイルを読み込む
	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", fmt.Errorf("メタデータファイル読み込みエラー: %w", err)
	}

	// JSONをパース
	var metadata Metadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return "", fmt.Errorf("JSONパースエラー: %w", err)
	}

	// IDが空の場合はエラー
	if metadata.ID == "" {
		return "", fmt.Errorf("メタデータにIDフィールドがありません")
	}

	return metadata.ID, nil
}
