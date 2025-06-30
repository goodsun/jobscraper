# カスタマイズガイド

## 新しいサイトに対応する方法

### 1. 自動検出を使う（JSON-LDがある場合）

多くの求人サイトはJSON-LD構造化データを持っているため、設定なしでも基本的な情報は取得できます：

```bash
go run src/universal-extractor.go "https://new-site.com/job/123" result.json
```

### 2. サイト固有の設定を作成

より詳細なデータを取得したい場合は、サイト設定ファイルを作成します：

#### ステップ1: サイトのHTML構造を調査

ブラウザの開発者ツールで要素を確認：
- 求人タイトルのセレクター
- 給与情報のセレクター
- 各フィールドの位置

#### ステップ2: 設定ファイルを作成

`configs/sites/your-site.json`:

```json
{
    "name": "your-site",
    "domain": "your-site.com",
    "selectors": {
        "name": "h1.job-title",
        "price": "div.salary span",
        "facility_name": "div.company-info h2",
        "area": "div.location-info",
        "occupation": "span.job-category",
        "contract": "span.employment-status",
        "detail": "div.job-details",
        "required_skill": "div.requirements ul",
        "holiday": "tr:contains('休日') td",
        "working_hours": "tr:contains('勤務時間') td",
        "dept": "div.department",
        "welfare_program": "div.benefits"
    }
}
```

#### ステップ3: 実行

```bash
go run src/universal-extractor.go "https://your-site.com/job/123" result.json your-site
```

### 3. セレクターの書き方

#### 基本的なセレクター
- `h1.job-title` - class="job-title"のh1要素
- `div#main-content` - id="main-content"のdiv要素
- `table.info td` - class="info"のtable内のtd要素

#### 高度なセレクター（jQuery風）
- `dt:contains('給与') + dd` - "給与"を含むdtの次のdd要素
- `tr:has(th:contains('勤務地')) td` - "勤務地"を含むthを持つtrのtd要素

### 4. デバッグ方法

1. まず設定なしで実行して、JSON-LDで何が取れるか確認
2. ブラウザで実際のページを開いて構造を確認
3. 設定ファイルを段階的に作成（最初は name と price だけなど）
4. 徐々にセレクターを追加

### 5. 実践例

#### Indeed の場合
```json
{
    "name": "indeed",
    "domain": "indeed.com",
    "selectors": {
        "name": "h1[data-testid='job-title']",
        "price": "span[data-testid='job-salary']",
        "facility_name": "div[data-testid='company-name']",
        "area": "div[data-testid='job-location']"
    }
}
```

#### マイナビ転職の場合
```json
{
    "name": "mynavi",
    "domain": "tenshoku.mynavi.jp",
    "selectors": {
        "name": "h1.jobname",
        "price": "div.salary",
        "facility_name": "p.companyname",
        "area": "div.workplace"
    }
}
```

## トラブルシューティング

### データが取得できない場合

1. **動的コンテンツの可能性**
   - JavaScriptで後から生成される内容は取得できません
   - browser-scraper.go の使用を検討

2. **セレクターの確認**
   - ブラウザのコンソールで確認：
   ```javascript
   document.querySelector('your-selector')
   ```

3. **文字エンコーディング**
   - 日本語が文字化けする場合は、サイトのエンコーディングを確認

### 共通パターン

多くの求人サイトは以下のパターンを使用：
- テーブル形式: `th`にラベル、`td`に値
- 定義リスト: `dt`にラベル、`dd`に値
- divブロック: ラベルと値が同じ親要素内

適切なセレクターを選択することで、ほとんどのサイトに対応できます。