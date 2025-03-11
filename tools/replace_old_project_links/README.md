# Replace Old Project Links

このツールは、指定されたディレクトリ内のMarkdownファイルに対して、CSVファイルで指定されたURLの置換を行います。




## 使用方法
1. qiitaに昔存在した、旧projects機能のURLとリダイレクト先をCSVでマップします。
1. 本ツールの実行

### コマンドライン引数

- `-dir`: 置換対象のMarkdownファイルが含まれるディレクトリを指定します。
- `-csv`: 置換対応表が記載されたCSVファイルを指定します。

### 実行例

```sh
go run [main.go](http://_vscodecontentref_/0) -dir /path/to/markdown/files -csv /path/to/replacements.csv
