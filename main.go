package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/qiita_export/models"
)

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

// TODO: 処理を書く
func execute(config *models.Config) error {

	// ループ
	//  perPage100件ずつAPIにリクエストを送る
	//  構造体にバインド
	//  メタデータと記事のコンテンツを保存するファイルの作成
	//  ファイルに値を入れる
	//  添付画像などのアセットを取得する（URLしか取れないので、URLを取ってそこにリクエストしてダウンロードする）
	//  CSVにパスを入れる
	//  page++

	return nil
}
