package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/qiita_export/models"
	"github.com/qiita_export/repository"
)

const (
	retryTimes = 5
)

var config *models.Config

func main() {
	outputDir := flag.String("dir", "output", "default value is 'output'")
	page := flag.Int("page", 1, "default value is 1")
	perPage := flag.Int("per_page", 100, "default value is 100")
	query := flag.String("query", "", "default value is empty")
	flag.Parse()

	// 時間計測用
	start := time.Now()

	// .envを環境変数にセットする
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// 設定
	config = models.NewConfig()
	if config.AccessToken == "" || config.Domain == "" {
		log.Fatalf("config required")
	}

	// 処理
	if err := execute(config, *outputDir, *page, *perPage, *query); err != nil {
		log.Fatalf("Error execute: %v", err)
	}

	fmt.Printf("実行時間: %f min, リクエスト数:%d", time.Since(start).Minutes(), repository.RequestCount)
}

func execute(config *models.Config, outputDir string, page, perPage int, query string) error {
	api := repository.NewQiitaAPI(config.Domain, config.AccessToken)

	for {
		params := fmt.Sprintf("page=%d&per_page=%d&query=%s", page, perPage, query)

		var articles []models.Article
		var requestErr error
		var total int
		// リトライ処理
		for range retryTimes {
			var err error
			articles, total, err = api.RequestArticles(params)
			if err != nil {
				requestErr = errors.Join(fmt.Errorf("failed to request page=%d, per_page=%d: %w", page, perPage, err))
				fmt.Printf("retry page=%d, error:%v\n", page, err)
				time.Sleep(5 * time.Second)
			} else {
				break
			}
		}

		if total <= 0 && requestErr != nil {
			return requestErr
		}

		// outputディレクトリの作成
		if err := os.MkdirAll(outputDir, 0777); err != nil {
			return err
		}

		// コメント, 絵文字の取得
		for _, v := range articles {
			comments, err := api.RequestComments(v.ID)
			if err != nil {
				return fmt.Errorf("コメントの取得に失敗しました: %w", err)
			}
			reactions, err := api.RequestArticleReactions(v.ID)
			if err != nil {
				return fmt.Errorf("絵文字リアクションの取得に失敗しました: %w", err)
			}
			v.Comments = comments
			v.EmojiReactions = reactions

			// mkdir
			groupDir := filepath.Join(outputDir, v.Group.Name)
			artDir := filepath.Join(groupDir, v.ID) // 記事名にSlashがある場合にエラーになるため、IDを採用
			if err := os.MkdirAll(artDir, 0777); err != nil {
				return err
			}

			if err := downloadArticleToLocal(&v, artDir); err != nil {
				return fmt.Errorf("記事のダウンロードに失敗しました: %w", err)
			}

			if err := api.DownloadArticleAssets(v.Body, artDir); err != nil {
				return fmt.Errorf("記事のアセットのダウンロードに失敗しました: %w", err)
			}
		}

		if page*perPage > total {
			break
		}

		// 進捗状況
		completed := min(page*perPage, total)
		progress := float64(completed) * 100 / float64(total)
		remainingPages := (max(0, total-page*perPage) + perPage - 1) / perPage

		fmt.Printf("Progress: %.1f%% (page=%d, remaining=%d)\n", progress, page, remainingPages)
		page++

		// API制限を考慮して、リクエスト間隔を空ける
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

func downloadArticleToLocal(art *models.Article, artDir string) error {
	fmt.Println(art.Title, strings.Repeat("=", 20))

	// ファイル名のサニタイズ
	sanitizedTitle := sanitizeFilename(art.Title)

	// メタデータの保存
	metadataPath := filepath.Join(artDir, sanitizedTitle+"_metadata.json")
	metadataJSON, err := json.MarshalIndent(art, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := os.WriteFile(metadataPath, metadataJSON, 0666); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	fmt.Println("メタデータの保存に成功しました")

	// Markdownファイルの保存
	mdPath := filepath.Join(artDir, sanitizedTitle+".md")
	if err := os.WriteFile(mdPath, []byte(art.Body), 0666); err != nil {
		return fmt.Errorf("failed to write markdown: %w", err)
	}
	fmt.Println("コンテンツの保存に成功しました")

	return nil
}

// ファイル名として使用できない文字をサニタイズする関数
func sanitizeFilename(filename string) string {
	// Windowsでも使用できるよう、一般的な禁止文字をすべて置換
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
