package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	// ルートディレクトリを指定
	root := "output"

	// IDと相対パスのマップを用意
	pathMap, err := createPathMap(root)
	if err != nil {
		log.Fatal(err)
	}

	replCount := 0

	// MDファイルを探索して置換
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			// 置換を実行
			newContent, err := replace(string(content), pathMap)
			if err != nil {
				return fmt.Errorf("failed to replace content in %s: %w", path, err)
			}

			// 変更があった場合のみ書き込み
			if newContent != string(content) {
				replCount++
				err = os.WriteFile(path, []byte(newContent), 0644)
				if err != nil {
					return fmt.Errorf("failed to write file %s: %w", path, err)
				}
				fmt.Printf("Updated: %s\n", path)
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(replCount)
}

func createPathMap(dir string) (map[string]string, error) {
	pathMap := make(map[string]string)

	// ディレクトリを再帰的に探索
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			parentDir := filepath.Base(filepath.Dir(path))
			if regexp.MustCompile(`^[0-9a-f]+$`).MatchString(parentDir) {
				relDir, err := filepath.Rel(dir, filepath.Dir(filepath.Dir(path)))
				if err != nil {
					return fmt.Errorf("failed to get relative path: %w", err)
				}

				// マークダウンで使用可能な相対パスを生成
				groupName := strings.Split(relDir, string(filepath.Separator))[0]
				mdName := strings.TrimSuffix(info.Name(), "_metadata.json")
				markdownRelativePath := fmt.Sprintf("/%s/%s/%s.md", groupName, parentDir, mdName)

				// 空白が含まれている場合、<>で囲まなければ認識されない
				if strings.Contains(markdownRelativePath, " ") {
					markdownRelativePath = fmt.Sprintf("<%s>", markdownRelativePath)
				}

				pathMap[parentDir] = markdownRelativePath
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pathMap, nil
}

// 文章を受け取り、正規表現でqiitaドメインのitemsを置換した値を取得する
func replace(body string, pathMap map[string]string) (string, error) {
	itemsRegexp := regexp.MustCompile(`https://qiita.com/[^/]+/items/([0-9a-z]+)(#[^)\s]*)?`)

	result := itemsRegexp.ReplaceAllStringFunc(body, func(s string) string {
		matches := itemsRegexp.FindStringSubmatch(s)
		if len(matches) < 2 {
			return s
		}

		id := matches[1]
		fragment := ""
		if len(matches) > 2 && matches[2] != "" {
			fragment = matches[2] // #が含まれた状態で取得される
		}

		if path, exists := pathMap[id]; exists {
			return fmt.Sprintf("%s%s", path, fragment)
		}

		return s
	})

	return result, nil
}
