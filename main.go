package main

import (
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
}

func main() {
	// .envを環境変数にセットする
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 設定
	config := Config{}

	// 処理
	if err := execute(&config); err != nil {
		log.Fatalf("Error execute: %v", err)
	}
}

func execute(config *Config) error {
	return nil
}
