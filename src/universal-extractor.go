package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// 汎用的なフィールド定義
type JobData struct {
	Name            string `json:"name"`
	Price           string `json:"price"`
	Area            string `json:"area"`
	Access          string `json:"access"`
	Address         string `json:"address"`
	City            string `json:"city"`
	Prefecture      string `json:"prefecture"`
	Contract        string `json:"contract"`
	Dept            string `json:"dept"`
	Detail          string `json:"detail"`
	FacilityName    string `json:"facility_name"`
	FacilityType    string `json:"facility_type"`
	Holiday         string `json:"holiday"`
	License         string `json:"license"`
	Occupation      string `json:"occupation"`
	Position        string `json:"position"`
	RequiredSkill   string `json:"required_skill"`
	StaffComment    string `json:"staff_comment"`
	Station         string `json:"station"`
	WelfareProgram  string `json:"welfare_program"`
	WorkingHours    string `json:"working_hours"`
	WorkingStyle    string `json:"working_style"`
	TitleOriginal   string `json:"title_original"`
}

// サイト別の抽出ルール
type SiteConfig struct {
	Name        string                       `json:"name"`
	Domain      string                       `json:"domain"`
	Encoding    string                       `json:"encoding"`
	Patterns    map[string]string           `json:"patterns"`
	Selectors   map[string]string           `json:"selectors"`
	Extractors  map[string]ExtractorConfig  `json:"extractors"`
}

type ExtractorConfig struct {
	Type     string `json:"type"`     // "selector", "regex", "json-ld"
	Value    string `json:"value"`    // CSS selector, regex pattern, etc.
	Attr     string `json:"attr"`     // attribute to extract (text, href, etc.)
	Index    int    `json:"index"`    // which match to use (default 0)
}

func fetchURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func convertEncoding(content string, encoding string) (string, error) {
	if encoding == "" || encoding == "utf-8" {
		return content, nil
	}
	
	var decoder transform.Transformer
	switch strings.ToLower(encoding) {
	case "shift_jis", "sjis":
		decoder = japanese.ShiftJIS.NewDecoder()
	case "euc-jp":
		decoder = japanese.EUCJP.NewDecoder()
	case "iso-2022-jp":
		decoder = japanese.ISO2022JP.NewDecoder()
	default:
		return content, nil // 未対応エンコーディングの場合はそのまま返す
	}
	
	result, _, err := transform.String(decoder, content)
	if err != nil {
		return content, err // エラーの場合は元の文字列を返す
	}
	
	return result, nil
}

func detectSite(url string) string {
	if strings.Contains(url, "kirara-support.jp") {
		return "kirara-support"
	} else if strings.Contains(url, "kyujiner.com") {
		return "kyujiner"
	} else if strings.Contains(url, "cme-pharmacist.jp") {
		return "cme-pharmacist"
	} else if strings.Contains(url, "th-agent.jp") {
		return "th-agent"
	} else if strings.Contains(url, "nursepower.co.jp") {
		return "nursepower"
	} else if strings.Contains(url, "nursejj.com") {
		return "nursejj"
	} else if strings.Contains(url, "yakumatch.com") {
		return "yakumatch"
	} else if strings.Contains(url, "supernurse.co.jp") {
		return "supernurse"
	} else if strings.Contains(url, "mc-nurse.net") {
		return "mc-nurse"
	} else if strings.Contains(url, "benesse-mcm.jp") {
		return "benesse-mcm"
	} else if strings.Contains(url, "kango-oshigoto.jp") {
		return "kango-oshigoto"
	} else if strings.Contains(url, "job.kiracare.jp") {
		return "kiracare"
	} else if strings.Contains(url, "pharmacareer.jp") {
		return "pharmacareer"
	} else if strings.Contains(url, "nurse-step.com") {
		return "nurse-step"
	}
	// 他のサイトの判定を追加
	return "default"
}

func loadSiteConfig(siteName string) (*SiteConfig, error) {
	configPath := filepath.Join("configs", "sites", siteName+".json")
	
	// カスタム設定ファイルを読み込む
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var config SiteConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}


func extractWithSelector(doc *goquery.Document, selector string) string {
	text := strings.TrimSpace(doc.Find(selector).First().Text())
	return normalizeSpaces(text)
}

// normalizeSpaces replaces multiple consecutive spaces with a single space,
// multiple consecutive newlines with double newlines, and removes link text
func normalizeSpaces(text string) string {
	// リンクテキストのパターンを削除
	linkPatterns := []string{
		`地図を見る`,
		`地図で見る`,
		`マップを見る`,
		`詳細を見る`,
		`詳しく見る`,
		`もっと見る`,
		`続きを見る`,
		`クリックして`,
		`こちらをクリック`,
		`詳細はこちら`,
		`▶`,
		`▼`,
		`▲`,
		`►`,
		`»`,
		`≫`,
		`>`,
	}
	
	// 各パターンを削除
	for _, pattern := range linkPatterns {
		linkRe := regexp.MustCompile(pattern)
		text = linkRe.ReplaceAllString(text, "")
	}
	
	// 3つ以上の改行を2つに置き換え
	newlineRe := regexp.MustCompile(`\n{3,}`)
	text = newlineRe.ReplaceAllString(text, "\n\n")
	
	// 複数の半角スペース（改行以外）を1つに置き換え
	spaceRe := regexp.MustCompile(`[^\S\n]{2,}`)
	text = spaceRe.ReplaceAllString(text, " ")
	
	// 前後の空白を削除
	text = strings.TrimSpace(text)
	
	return text
}

func extractData(htmlContent string, config *SiteConfig) (*JobData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	data := &JobData{}

	// JSON-LD extraction (共通)
	extractJSONLD(htmlContent, data)

	// セレクターベースの抽出（ハイブリッド方式：JSON-LDとセレクターを組み合わせ）
	if config.Selectors != nil {
		// 重要な基本情報は常にセレクターを優先（既存サイト互換性のため）
		if selector, ok := config.Selectors["name"]; ok {
			if selectorValue := extractWithSelector(doc, selector); selectorValue != "" {
				data.Name = selectorValue
				data.TitleOriginal = selectorValue
			}
		}
		if selector, ok := config.Selectors["price"]; ok {
			if selectorValue := extractWithSelector(doc, selector); selectorValue != "" {
				data.Price = selectorValue
			}
		}
		
		// その他の情報はJSON-LDを優先し、取得できない場合のみセレクターを使用
		if selector, ok := config.Selectors["facility_name"]; ok && data.FacilityName == "" {
			data.FacilityName = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["area"]; ok && data.Area == "" {
			data.Area = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["access"]; ok && data.Access == "" {
			data.Access = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["occupation"]; ok && data.Occupation == "" {
			data.Occupation = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["contract"]; ok && data.Contract == "" {
			data.Contract = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["staff_comment"]; ok && data.StaffComment == "" {
			data.StaffComment = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["dept"]; ok && data.Dept == "" {
			data.Dept = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["detail"]; ok && data.Detail == "" {
			data.Detail = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["facility_type"]; ok && data.FacilityType == "" {
			data.FacilityType = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["holiday"]; ok && data.Holiday == "" {
			data.Holiday = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["license"]; ok && data.License == "" {
			data.License = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["required_skill"]; ok && data.RequiredSkill == "" {
			data.RequiredSkill = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["station"]; ok && data.Station == "" {
			data.Station = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["welfare_program"]; ok && data.WelfareProgram == "" {
			data.WelfareProgram = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["working_hours"]; ok && data.WorkingHours == "" {
			data.WorkingHours = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["working_style"]; ok && data.WorkingStyle == "" {
			data.WorkingStyle = extractWithSelector(doc, selector)
		}
	}

	// 住所から都道府県と市区町村を抽出
	if data.Address != "" {
		extractLocationInfo(data)
	} else if data.Area != "" {
		data.Address = data.Area
		extractLocationInfo(data)
	}

	return data, nil
}

func extractJSONLD(htmlContent string, data *JobData) {
	// JSON-LD スキーマの抽出
	structuredDataRegex := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>(.*?)</script>`)
	matches := structuredDataRegex.FindAllStringSubmatch(htmlContent, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			// 単一オブジェクトを試す
			var singleJsonData map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &singleJsonData); err == nil {
				if singleJsonData["@type"] == "JobPosting" {
					extractFromJobPosting(singleJsonData, data)
				}
			} else {
				// 配列形式を試す
				var jsonDataArray []map[string]interface{}
				if err := json.Unmarshal([]byte(match[1]), &jsonDataArray); err == nil {
					for _, item := range jsonDataArray {
						if item["@type"] == "JobPosting" {
							extractFromJobPosting(item, data)
						}
					}
				}
			}
		}
	}
}


func extractFromJobPosting(item map[string]interface{}, data *JobData) {
	// タイトル
	if title, ok := item["title"].(string); ok && data.Name == "" {
		data.Name = normalizeSpaces(title)
		data.TitleOriginal = normalizeSpaces(title)
	}
	if desc, ok := item["description"].(string); ok && data.Name == "" {
		lines := strings.Split(desc, "<br>")
		if len(lines) > 0 {
			data.Name = normalizeSpaces(strings.TrimSpace(lines[0]))
			data.TitleOriginal = data.Name
		}
	}
	
	// 給与
	if baseSalary, ok := item["baseSalary"].(map[string]interface{}); ok {
		if data.Price == "" {
			if value, ok := baseSalary["value"].(map[string]interface{}); ok {
				// float64とintの両方をチェック
				if minVal, ok := value["minValue"].(float64); ok {
					if maxVal, ok := value["maxValue"].(float64); ok {
						data.Price = fmt.Sprintf("月収 %.0f〜%.0f円", minVal, maxVal)
					}
				} else if minValInt, ok := value["minValue"].(int); ok {
					if maxValInt, ok := value["maxValue"].(int); ok {
						data.Price = fmt.Sprintf("月収 %d〜%d円", minValInt, maxValInt)
					}
				}
			}
		}
	}
	
	// 勤務地
	if jobLocation, ok := item["jobLocation"].(map[string]interface{}); ok {
		if address, ok := jobLocation["address"].(map[string]interface{}); ok {
			if region, ok := address["addressRegion"].(string); ok && data.Prefecture == "" {
				data.Prefecture = normalizeSpaces(region)
			}
			if locality, ok := address["addressLocality"].(string); ok && data.City == "" {
				data.City = normalizeSpaces(locality)
			}
			if street, ok := address["streetAddress"].(string); ok && data.Address == "" {
				data.Address = normalizeSpaces(fmt.Sprintf("%s%s%s", data.Prefecture, data.City, street))
			}
			if data.Area == "" {
				data.Area = normalizeSpaces(data.Prefecture + data.City)
			}
		}
	}
	
	// 施設名
	if org, ok := item["hiringOrganization"].(map[string]interface{}); ok {
		if name, ok := org["name"].(string); ok && data.FacilityName == "" {
			data.FacilityName = normalizeSpaces(name)
		}
	}
	
	// 職種カテゴリー
	if occCategory, ok := item["occupationalCategory"].(string); ok && data.Occupation == "" {
		data.Occupation = normalizeSpaces(occCategory)
	}
	
	// 雇用形態
	if empType, ok := item["employmentType"].(string); ok && data.Contract == "" {
		switch empType {
		case "FULL_TIME":
			data.Contract = "正社員(常勤)"
		case "PART_TIME":
			data.Contract = "非常勤"
		case "CONTRACT":
			data.Contract = "契約社員"
		}
	}
	
	// 勤務時間
	if workHours, ok := item["workHours"].(string); ok && data.WorkingHours == "" {
		data.WorkingHours = normalizeSpaces(workHours)
	}
	
	// 必要資格
	if qualifications, ok := item["qualifications"].(string); ok && data.License == "" {
		data.License = normalizeSpaces(qualifications)
	}
	
	// 仕事内容
	if responsibilities, ok := item["responsibilities"].(string); ok && data.Detail == "" {
		data.Detail = normalizeSpaces(responsibilities)
	}
	
	// 福利厚生
	if benefits, ok := item["jobBenefits"].(string); ok && data.WelfareProgram == "" {
		data.WelfareProgram = normalizeSpaces(benefits)
	}
	
	// descriptionから詳細情報を抽出
	if desc, ok := item["description"].(string); ok {
		extractFromDescription(desc, data)
	}
}

// descriptionから詳細情報を抽出する関数
func extractFromDescription(desc string, data *JobData) {
	// 雇用形態の抽出 (常勤、非常勤、正社員等)
	if data.Contract == "" {
		if strings.Contains(desc, "常勤") {
			if strings.Contains(desc, "夜勤有り") || strings.Contains(desc, "夜勤あり") {
				data.Contract = "正社員(常勤・夜勤有り)"
			} else {
				data.Contract = "正社員(常勤)"
			}
		} else if strings.Contains(desc, "非常勤") {
			data.Contract = "非常勤"
		} else if strings.Contains(desc, "正社員") {
			data.Contract = "正社員"
		}
	}
	
	// 配属先の抽出
	if data.Position == "" {
		if strings.Contains(desc, "配属先：病棟") || strings.Contains(desc, "病棟") {
			data.Position = "病棟"
		} else if strings.Contains(desc, "配属先：外来") || strings.Contains(desc, "外来") {
			data.Position = "外来"
		} else if strings.Contains(desc, "配属先：手術室") || strings.Contains(desc, "手術室") {
			data.Position = "手術室"
		}
	}
	
	// 診療科目の抽出
	if data.Dept == "" {
		// 診療科目のパターンを探す
		deptRegex := regexp.MustCompile(`診療科目[：:]\s*([^<\n]+)`)
		if matches := deptRegex.FindStringSubmatch(desc); len(matches) > 1 {
			data.Dept = normalizeSpaces(strings.TrimSpace(matches[1]))
		}
	}
	
	// 施設形態の抽出
	if data.FacilityType == "" {
		facilityRegex := regexp.MustCompile(`施設形態[：:]\s*([^<\n]+)`)
		if matches := facilityRegex.FindStringSubmatch(desc); len(matches) > 1 {
			data.FacilityType = normalizeSpaces(strings.TrimSpace(matches[1]))
		}
	}
	
	// 勤務形態の抽出（2交替、3交替等）
	if data.WorkingStyle == "" {
		if strings.Contains(desc, "2交替") || strings.Contains(desc, "二交替") {
			data.WorkingStyle = "2交替"
		} else if strings.Contains(desc, "3交替") || strings.Contains(desc, "三交替") {
			data.WorkingStyle = "3交替"
		}
	}
}

func extractLocationInfo(data *JobData) {
	// 都道府県の抽出
	prefectureRegex := regexp.MustCompile(`^([^都道府県]+[都道府県])`)
	if matches := prefectureRegex.FindStringSubmatch(data.Address); len(matches) > 1 {
		data.Prefecture = matches[1]
	}
	
	// 市区町村の抽出
	cityRegex := regexp.MustCompile(`[都道府県]([^区市町村]+[区市町村])`)
	if matches := cityRegex.FindStringSubmatch(data.Address); len(matches) > 1 {
		data.City = matches[1]
	}
	
	if data.Area == "" && data.Prefecture != "" && data.City != "" {
		data.Area = data.Prefecture + data.City
	}
}

func listConfigs() {
	configDir := filepath.Join("configs", "sites")
	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		fmt.Printf("Error reading config directory: %v\n", err)
		return
	}
	
	fmt.Println("Available Site Configurations")
	fmt.Println("=============================")
	fmt.Println()
	
	var configs []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			configName := strings.TrimSuffix(file.Name(), ".json")
			configs = append(configs, configName)
		}
	}
	
	sort.Strings(configs)
	
	for _, config := range configs {
		// 設定ファイルを読み込んでドメインを取得
		configPath := filepath.Join(configDir, config+".json")
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			fmt.Printf("%-20s (error reading config)\n", config)
			continue
		}
		
		var siteConfig SiteConfig
		if err := json.Unmarshal(data, &siteConfig); err != nil {
			fmt.Printf("%-20s (error parsing config)\n", config)
			continue
		}
		
		domain := siteConfig.Domain
		if domain == "" {
			domain = "no domain specified"
		}
		fmt.Printf("%-20s - %s\n", config, domain)
	}
	
	fmt.Println()
	fmt.Println("Usage: universal-extractor --config <name> <url>")
}

func showHelp() {
	fmt.Println("Universal Job Data Extractor")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  universal-extractor [options] <url> [output_file]")
	fmt.Println("  universal-extractor -h | --help")
	fmt.Println("  universal-extractor --list-configs")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <url>          - 求人詳細ページのURL")
	fmt.Println("  [output_file]  - 出力ファイル名（省略時は標準出力）")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      - このヘルプメッセージを表示")
	fmt.Println("  --config <name> - サイト設定を指定（省略時は自動検出）")
	fmt.Println("  --list-configs  - 利用可能な設定ファイル一覧を表示")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # 標準出力に表示（自動サイト検出）")
	fmt.Println("  universal-extractor https://example.com/job/123")
	fmt.Println()
	fmt.Println("  # ファイルに保存（自動サイト検出）")
	fmt.Println("  universal-extractor https://example.com/job/123 output.json")
	fmt.Println()
	fmt.Println("  # 特定のサイト設定を使用")
	fmt.Println("  universal-extractor --config custom-site https://example.com/job/123")
	fmt.Println()
	fmt.Println("  # 利用可能な設定を確認")
	fmt.Println("  universal-extractor --list-configs")
	fmt.Println()
	fmt.Println("Supported Sites:")
	fmt.Println("  - benesse-mcm.jp")
	fmt.Println("  - cme-pharmacist.jp")
	fmt.Println("  - job.kiracare.jp")
	fmt.Println("  - kango-oshigoto.jp")
	fmt.Println("  - kirara-support.jp")
	fmt.Println("  - kyujiner.com")
	fmt.Println("  - mc-nurse.net")
	fmt.Println("  - nurse-step.com")
	fmt.Println("  - nursejj.com")
	fmt.Println("  - nursepower.co.jp")
	fmt.Println("  - pharmacareer.jp")
	fmt.Println("  - supernurse.co.jp")
	fmt.Println("  - th-agent.jp")
	fmt.Println("  - yakumatch.com")
}

func main() {
	var url string
	var outputFile string
	var siteName string
	var configSpecified bool

	// 引数解析
	args := os.Args[1:]
	i := 0
	
	for i < len(args) {
		arg := args[i]
		
		// ヘルプオプション
		if arg == "-h" || arg == "--help" {
			showHelp()
			os.Exit(0)
		}
		
		// 設定一覧オプション
		if arg == "--list-configs" {
			listConfigs()
			os.Exit(0)
		}
		
		// 設定ファイルオプション
		if arg == "--config" {
			if i+1 >= len(args) {
				fmt.Println("Error: --config requires a site name")
				os.Exit(1)
			}
			siteName = args[i+1]
			configSpecified = true
			i += 2
			continue
		}
		
		// URL（最初の非オプション引数）
		if url == "" && !strings.HasPrefix(arg, "-") {
			url = arg
		} else if outputFile == "" && !strings.HasPrefix(arg, "-") {
			// 出力ファイル（2番目の非オプション引数）
			outputFile = arg
		}
		
		i++
	}
	
	// URL必須チェック
	if url == "" {
		showHelp()
		os.Exit(1)
	}
	
	// サイト設定の自動検出（--configが指定されていない場合）
	if !configSpecified {
		siteName = detectSite(url)
	}
	
	fmt.Printf("Using site configuration: %s\n", siteName)
	
	// サイト設定を読み込む
	config, err := loadSiteConfig(siteName)
	if err != nil {
		fmt.Printf("Warning: Could not load site config for %s, using generic extraction\n", siteName)
		config = &SiteConfig{Name: "default"}
	}

	// URLからHTMLを取得
	fmt.Printf("Fetching data from URL: %s\n", url)
	htmlContent, err := fetchURL(url)
	if err != nil {
		log.Fatal("Error fetching URL:", err)
	}

	// エンコーディング変換
	if config.Encoding != "" {
		convertedContent, err := convertEncoding(htmlContent, config.Encoding)
		if err != nil {
			fmt.Printf("Warning: Failed to convert encoding from %s: %v\n", config.Encoding, err)
		} else {
			htmlContent = convertedContent
			fmt.Printf("Converted encoding from %s to UTF-8\n", config.Encoding)
		}
	}

	// データを抽出
	jobData, err := extractData(htmlContent, config)
	if err != nil {
		log.Fatal("Error extracting data:", err)
	}

	// JSONに変換
	jsonData, err := json.MarshalIndent(jobData, "", "    ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}

	// 出力先が指定されていない場合は標準出力に表示
	if outputFile == "" {
		fmt.Println(string(jsonData))
	} else {
		// ファイルに保存
		err = ioutil.WriteFile(outputFile, jsonData, 0644)
		if err != nil {
			log.Fatal("Error writing output file:", err)
		}
		fmt.Printf("Data extracted successfully and saved to %s\n", outputFile)
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
