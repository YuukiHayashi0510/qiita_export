package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/qiita_export/models"
	"github.com/qiita_export/repository"
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

	repo := repository.ArticleMetadata{}

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

		// マークダウンファイルを読み込む
		mdContent, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("マークダウンファイル読み込みエラー: %w", err)
		}

		// メタデータから記事を取得
		article, err := repo.GetArticle(metadataPath)
		if err != nil {
			return err
		}

		// メタデータをマークダウンに追加
		updatedContent := string(mdContent) + createMetadataForMarkdown(article)
		// メタデータファイルへのリンクをマークダウンに追加
		updatedContent += createFileLink(baseName + "_metadata.json")

		// ファイル更新
		if err = os.WriteFile(path, []byte(updatedContent), 0644); err != nil {
			return fmt.Errorf("ファイル書き込みエラー: %w", err)
		}
		fmt.Printf("メタデータ追加完了 title=%s", filepath.Base(path))

		return nil
	})

	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}
}

func createMetadataForMarkdown(article *models.Article) string {
	sep := "\n---\n"
	codeBlock := func(value string) string {
		return fmt.Sprintf("```\n%s\n```\n", value)
	}

	idKeyValue := "Qiitaの記事ID: " + article.ID

	return sep + codeBlock(idKeyValue)
}

func createFileLink(metadataFileName string) string {
	// 半角全角スペースが含まれていない場合はそのまま付与
	if !strings.Contains(metadataFileName, " ") && !strings.Contains(metadataFileName, "　") {
		return sprintFileLink(metadataFileName, metadataFileName)
	}
	// ()が含まれていない場合はそのまま付与
	if !strings.Contains(metadataFileName, "(") && strings.Contains(metadataFileName, ")") {
		return sprintFileLink(metadataFileName, metadataFileName)
	}

	// ファイル名にマークダウンでリンクとして認識されない記号がある場合、<>で囲む
	linkValue := fmt.Sprintf("<%s>", metadataFileName)
	return sprintFileLink(metadataFileName, linkValue)
}

func sprintFileLink(title, link string) string {
	return fmt.Sprintf("\n[%s](%s)\n", title, link)
}
