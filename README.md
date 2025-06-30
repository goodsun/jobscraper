# Job Data Extractor

求人サイトから自動的にデータを抽出するツール

## ディレクトリ構造

```
.
├── src/                    # ソースコード
│   ├── job-extractor.go    # メインの抽出ツール（推奨）
│   ├── scraper.go          # XPathベースのスクレイパー
│   └── browser-scraper.go  # ブラウザレンダリング版（開発中）
├── format/                 # フォーマット定義
│   ├── format.json         # 空のテンプレート
│   └── sample*.json        # サンプルXPath設定
├── configs/                # XPath設定ファイル
├── examples/               # テスト用HTMLファイル
├── output/                 # 出力されたJSONファイル
└── go.mod                  # Go依存関係

## 使い方

### 1. メインツール（推奨）

URLから直接データを抽出：
```bash
go run src/job-extractor.go "https://example.com/job/123" output.json
```

HTMLファイルから抽出：
```bash
go run src/job-extractor.go examples/job.html output.json
```

### 2. ビルドして使用

```bash
# ビルド
go build -o job-extractor src/job-extractor.go

# 実行
./job-extractor "https://example.com/job/123" result.json
```

## 機能

- URL/HTMLファイルの両方に対応
- JSON-LDスキーマの自動解析
- DOM構造の自動認識
- 完全なデータ抽出（全フィールド対応）

## 出力フォーマット

```json
{
    "name": "求人タイトル",
    "price": "給与",
    "area": "エリア",
    "facility_name": "施設名",
    "dept": "診療科目",
    "occupation": "職種",
    ...
}
```