package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type XPathConfig struct {
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

type ScrapedData struct {
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

func extractByXPath(ctx context.Context, xpath string) string {
	if xpath == "" {
		return ""
	}

	var result []string
	err := chromedp.Run(ctx,
		chromedp.Text(xpath, &result, chromedp.BySearch),
	)
	if err != nil || len(result) == 0 {
		// Try alternative approach
		var text string
		err = chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf(`
				try {
					var xpathResult = document.evaluate('%s', document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null);
					if (xpathResult.singleNodeValue) {
						return xpathResult.singleNodeValue.textContent.trim();
					}
					return "";
				} catch (e) {
					return "";
				}
			`, strings.ReplaceAll(xpath, "'", "\\'")), &text),
		)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(result[0])
}

func scrapeData(url string, config *XPathConfig) (*ScrapedData, error) {
	// Create context with timeout
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	
	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Navigate to the page and wait for it to load
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(3*time.Second), // Wait for dynamic content to load
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate: %v", err)
	}

	fmt.Println("Page loaded, extracting data...")

	// Extract data using XPaths
	data := &ScrapedData{}
	
	// Extract each field with debug info
	fmt.Printf("Extracting name with XPath: %s\n", config.Name)
	data.Name = extractByXPath(ctx, config.Name)
	fmt.Printf("  Result: %s\n", data.Name)
	
	fmt.Printf("Extracting price with XPath: %s\n", config.Price)
	data.Price = extractByXPath(ctx, config.Price)
	fmt.Printf("  Result: %s\n", data.Price)
	
	data.Area = extractByXPath(ctx, config.Area)
	data.Access = extractByXPath(ctx, config.Access)
	data.Address = extractByXPath(ctx, config.Address)
	data.City = extractByXPath(ctx, config.City)
	data.Prefecture = extractByXPath(ctx, config.Prefecture)
	data.Contract = extractByXPath(ctx, config.Contract)
	data.Dept = extractByXPath(ctx, config.Dept)
	data.Detail = extractByXPath(ctx, config.Detail)
	data.FacilityName = extractByXPath(ctx, config.FacilityName)
	data.FacilityType = extractByXPath(ctx, config.FacilityType)
	data.Holiday = extractByXPath(ctx, config.Holiday)
	data.License = extractByXPath(ctx, config.License)
	data.Occupation = extractByXPath(ctx, config.Occupation)
	data.Position = extractByXPath(ctx, config.Position)
	data.RequiredSkill = extractByXPath(ctx, config.RequiredSkill)
	data.StaffComment = extractByXPath(ctx, config.StaffComment)
	data.Station = extractByXPath(ctx, config.Station)
	data.WelfareProgram = extractByXPath(ctx, config.WelfareProgram)
	data.WorkingHours = extractByXPath(ctx, config.WorkingHours)
	data.WorkingStyle = extractByXPath(ctx, config.WorkingStyle)
	data.TitleOriginal = extractByXPath(ctx, config.TitleOriginal)

	return data, nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run browser-scraper.go <url> <xpath_config.json> <output.json>")
		fmt.Println("Example: go run browser-scraper.go https://example.com/job xpath_config.json result.json")
		os.Exit(1)
	}

	url := os.Args[1]
	configFile := os.Args[2]
	outputFile := os.Args[3]

	// Read XPath configuration
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	var config XPathConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatal("Error parsing config JSON:", err)
	}

	// Scrape data
	fmt.Printf("Starting browser-based scraping for: %s\n", url)
	scrapedData, err := scrapeData(url, &config)
	if err != nil {
		log.Fatal("Error scraping data:", err)
	}

	// Output to JSON
	jsonData, err := json.MarshalIndent(scrapedData, "", "    ")
	if err != nil {
		log.Fatal("Error marshaling JSON:", err)
	}

	err = ioutil.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		log.Fatal("Error writing output file:", err)
	}

	fmt.Printf("Data scraped successfully and saved to %s\n", outputFile)
}