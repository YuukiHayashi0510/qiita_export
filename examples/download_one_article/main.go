package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/qiita_export/models"
)

const (
	outputDir = "output"
)

// 一つだけ記事を取得し、ダウンロードしてみる
func main() {
	// .envを環境変数にセットする
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// 設定
	config := models.NewConfig()
	if config.AccessToken == "" || config.Domain == "" {
		log.Fatalf("config required")
	}

	// 処理
	if err := execute(config); err != nil {
		log.Fatalf("Error execute: %v", err)
	}
}

func execute(config *models.Config) error {
	url := fmt.Sprintf("https://%s/api/v2/items?page=%d&per_page=%d", config.Domain, 1, 1)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	token := fmt.Sprintf("Bearer %s", config.AccessToken)
	req.Header.Set("Authorization", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request api: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}

	var articles []*models.Article
	if err := json.NewDecoder(res.Body).Decode(&articles); err != nil {
		return err
	}

	// outputディレクトリの作成
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return err
	}

	// ダウンロード
	for _, v := range articles {
		fmt.Println(v.Title, strings.Repeat("=", 20))

		// mkdir
		groupDir := filepath.Join(outputDir, v.Group.Name)
		artDir := filepath.Join(groupDir, v.Title)
		if err := os.MkdirAll(artDir, 0777); err != nil {
			return err
		}

		// メタデータの保存
		metadataPath := filepath.Join(artDir, "metadata.json")
		metadataJSON, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		if err := os.WriteFile(metadataPath, metadataJSON, 0666); err != nil {
			return fmt.Errorf("failed to write metadata: %w", err)
		}
		fmt.Println("メタデータの保存に成功しました")

		// Markdownファイルの保存
		mdPath := filepath.Join(artDir, v.Title+".md")
		if err := os.WriteFile(mdPath, []byte(v.Body), 0666); err != nil {
			return fmt.Errorf("failed to write markdown: %w", err)
		}
		fmt.Println("コンテンツの保存に成功しました")
	}

	return nil
}
