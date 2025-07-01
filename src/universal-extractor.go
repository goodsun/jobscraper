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
	return strings.TrimSpace(doc.Find(selector).First().Text())
}

func extractData(htmlContent string, config *SiteConfig) (*JobData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	data := &JobData{}

	// JSON-LD extraction (共通)
	extractJSONLD(htmlContent, data)

	// セレクターベースの抽出
	if config.Selectors != nil {
		if selector, ok := config.Selectors["name"]; ok {
			data.Name = extractWithSelector(doc, selector)
			data.TitleOriginal = data.Name
		}
		if selector, ok := config.Selectors["price"]; ok {
			data.Price = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["facility_name"]; ok {
			data.FacilityName = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["area"]; ok {
			data.Area = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["access"]; ok {
			data.Access = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["occupation"]; ok {
			data.Occupation = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["contract"]; ok {
			data.Contract = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["staff_comment"]; ok {
			data.StaffComment = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["dept"]; ok {
			data.Dept = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["detail"]; ok {
			data.Detail = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["facility_type"]; ok {
			data.FacilityType = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["holiday"]; ok {
			data.Holiday = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["license"]; ok {
			data.License = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["required_skill"]; ok {
			data.RequiredSkill = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["station"]; ok {
			data.Station = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["welfare_program"]; ok {
			data.WelfareProgram = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["working_hours"]; ok {
			data.WorkingHours = extractWithSelector(doc, selector)
		}
		if selector, ok := config.Selectors["working_style"]; ok {
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
			var jsonData []map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &jsonData); err == nil {
				for _, item := range jsonData {
					if item["@type"] == "JobPosting" {
						extractFromJobPosting(item, data)
					}
				}
			}
		}
	}
}

func extractFromJobPosting(item map[string]interface{}, data *JobData) {
	// タイトル
	if desc, ok := item["description"].(string); ok && data.Name == "" {
		lines := strings.Split(desc, "<br>")
		if len(lines) > 0 {
			data.Name = strings.TrimSpace(lines[0])
			data.TitleOriginal = data.Name
		}
	}
	
	// 給与
	if baseSalary, ok := item["baseSalary"].(map[string]interface{}); ok && data.Price == "" {
		if value, ok := baseSalary["value"].(map[string]interface{}); ok {
			if min, ok := value["minValue"].(string); ok {
				if max, ok := value["maxValue"].(string); ok {
					data.Price = fmt.Sprintf("年収 %s〜%s円", min, max)
				}
			}
		}
	}
	
	// 勤務地
	if jobLocation, ok := item["jobLocation"].(map[string]interface{}); ok {
		if address, ok := jobLocation["address"].(map[string]interface{}); ok {
			if region, ok := address["addressRegion"].(string); ok && data.Prefecture == "" {
				data.Prefecture = region
			}
			if locality, ok := address["addressLocality"].(string); ok && data.City == "" {
				data.City = locality
			}
			if street, ok := address["streetAddress"].(string); ok && data.Address == "" {
				data.Address = fmt.Sprintf("%s%s%s", data.Prefecture, data.City, street)
			}
			if data.Area == "" {
				data.Area = data.Prefecture + data.City
			}
		}
	}
	
	// 施設名
	if org, ok := item["hiringOrganization"].(map[string]interface{}); ok {
		if name, ok := org["name"].(string); ok && data.FacilityName == "" {
			data.FacilityName = name
		}
	}
	
	// 雇用形態
	if empType, ok := item["employmentType"].(string); ok && data.Contract == "" {
		if empType == "FULL_TIME" {
			data.Contract = "正社員(常勤)"
		}
	}
	
	// 職種
	if title, ok := item["title"].(string); ok && data.Occupation == "" {
		data.Occupation = title
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
	fmt.Println("  - kango.kyujiner.com")
	fmt.Println("  - kirara-support.jp")
	fmt.Println("  - mc-nurse.net")
	fmt.Println("  - nurse-pw.jp")
	fmt.Println("  - nursejj.com")
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