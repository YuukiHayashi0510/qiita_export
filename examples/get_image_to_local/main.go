package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/qiita_export/models"
)

const (
	body = ""
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	config := models.NewConfig()
	imgRegexp := regexp.MustCompile(fmt.Sprintf(`https://%s/.+\.png`, config.Domain))

	var (
		rerr  error
		count int
	)

	_ = imgRegexp.ReplaceAllStringFunc(body, func(s string) string {
		if rerr != nil {
			return s
		}

		count++
		f, err := os.Create(filepath.Join("", path.Base(s)))
		if err != nil {
			rerr = err
			return s
		}

		req, err := http.NewRequest(http.MethodGet, s, nil)
		if err != nil {
			rerr = err
			return s
		}

		token := fmt.Sprintf("Bearer %s", config.AccessToken)
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			rerr = err
			return s
		}
		defer resp.Body.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			rerr = err
			return s
		}

		if err := f.Close(); err != nil {
			rerr = err
			return s
		}

		return s
	})
	if rerr != nil {
		log.Fatal(rerr)
	}

	fmt.Println(count, rerr)
}
