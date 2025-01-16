package models

// Qiita Tag
type Tag struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}
