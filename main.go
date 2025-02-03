package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/qiita_export/models"
)

const (
	outputDir      = "output"
	defaultPage    = 1
	defaultPerPage = 100
	retryTimes     = 5
	sleepTime      = 100 * time.Millisecond
)

var config *models.Config

func main() {
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
	if err := execute(config); err != nil {
		log.Fatalf("Error execute: %v", err)
	}

	fmt.Printf("実行時間: %f min", time.Now().Sub(start).Minutes())
}

func execute(config *models.Config) error {
	page := defaultPage
	perPage := defaultPerPage

	for {
		url := fmt.Sprintf("https://%s/api/v2/items?page=%d&per_page=%d", config.Domain, page, perPage)
		token := fmt.Sprintf("Bearer %s", config.AccessToken)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", token)

		total := -1
		var requestErr error
		for i := 0; i < retryTimes; i++ {
			t, err := request(req)
			if err != nil {
				requestErr = errors.Join(fmt.Errorf("failed to request page=%d, per_page=%d: %w", page, perPage, err))
				fmt.Printf("retry page=%d\n", page)
				time.Sleep(5 * time.Second)
			} else {
				total = t
				break
			}
		}
		if total == -1 && requestErr != nil {
			return requestErr
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
	}

	return nil
}

func request(req *http.Request) (int, error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, fmt.Errorf("failed to request api: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return -1, errors.New(res.Status)
	}

	var articles []*models.Article
	if err := json.NewDecoder(res.Body).Decode(&articles); err != nil {
		return -1, err
	}

	// outputディレクトリの作成
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return -1, err
	}

	// ダウンロード
	for _, v := range articles {
		if err := downloadArticle(v); err != nil {
			return -1, err
		}
	}

	total, err := strconv.Atoi(res.Header.Get("Total-Count"))
	if err != nil {
		return -1, err
	}

	return total, nil
}

func downloadArticle(art *models.Article) error {
	fmt.Println(art.Title, strings.Repeat("=", 20))

	// mkdir
	groupDir := filepath.Join(outputDir, art.Group.Name)
	artDir := filepath.Join(groupDir, art.ID) // 記事名にSlashがある場合にエラーになるため、IDを採用
	if err := os.MkdirAll(artDir, 0777); err != nil {
		return err
	}

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

	// 画像のダウンロード
	if err := downloadArticleAssets(art.Body, artDir); err != nil {
		return err
	}

	return nil
}

// 添付画像のダウンロード
func downloadArticleAssets(body, artDir string) (retErr error) {
	assetRegexp := regexp.MustCompile(`https://qiita\.com/files/[0-9a-z-]+\.[^]\s"'<>)]+`)

	count := 0
	_ = assetRegexp.ReplaceAllStringFunc(body, func(s string) string {
		count++
		if retErr != nil {
			return s
		}

		f, err := os.Create(filepath.Join(artDir, path.Base(s)))
		if err != nil {
			retErr = err
			return s
		}

		req, err := http.NewRequest(http.MethodGet, s, nil)
		if err != nil {
			retErr = err
			return s
		}

		token := fmt.Sprintf("Bearer %s", config.AccessToken)
		req.Header.Set("Authorization", token)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			retErr = err
			return s
		}
		defer res.Body.Close()

		if res.StatusCode == 403 {
			fmt.Println("403:", artDir, s)
		}

		if _, err := io.Copy(f, res.Body); err != nil {
			retErr = err
			return s
		}

		if err := f.Close(); err != nil {
			retErr = err
			return s
		}

		// レート制限で403になってしまうため待機時間を設ける
		time.Sleep(sleepTime)

		return s
	})

	fmt.Println("total assets", count)

	return
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
