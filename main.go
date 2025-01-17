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

	"github.com/joho/godotenv"
	"github.com/qiita_export/models"
)

const (
	outputDir      = "output"
	defaultPerPage = 1
)

var config *models.Config

func main() {
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
}

func execute(config *models.Config) error {
	page := 1
	perPage := defaultPerPage

	for {
		url := fmt.Sprintf("https://%s/api/v2/items?page=%d&per_page=%d", config.Domain, page, perPage)
		token := fmt.Sprintf("Bearer %s", config.AccessToken)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", token)

		total, err := request(req)
		if err != nil {
			return fmt.Errorf("failed to request page=%d, per_page=%d: %w", page, perPage, err)
		}

		if page*perPage > total {
			break
		}

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
	artDir := filepath.Join(groupDir, art.Title)
	if err := os.MkdirAll(artDir, 0777); err != nil {
		return err
	}

	// メタデータの保存
	metadataPath := filepath.Join(artDir, "metadata.json")
	metadataJSON, err := json.MarshalIndent(art, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := os.WriteFile(metadataPath, metadataJSON, 0666); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	fmt.Println("メタデータの保存に成功しました")

	// Markdownファイルの保存
	mdPath := filepath.Join(artDir, art.Title+".md")
	if err := os.WriteFile(mdPath, []byte(art.Body), 0666); err != nil {
		return fmt.Errorf("failed to write markdown: %w", err)
	}
	fmt.Println("コンテンツの保存に成功しました")

	// 画像のダウンロード
	if err := downloadArticleImages(art.Body, artDir); err != nil {
		return err
	}

	return nil
}

// 添付画像のダウンロード
func downloadArticleImages(body, artDir string) (retErr error) {
	imgRegexp := regexp.MustCompile(fmt.Sprintf(`https://%s/.+\.png`, config.Domain))

	_ = imgRegexp.ReplaceAllStringFunc(body, func(s string) string {
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

		if _, err := io.Copy(f, res.Body); err != nil {
			retErr = err
			return s
		}

		if err := f.Close(); err != nil {
			retErr = err
			return s
		}

		return s
	})

	return
}
