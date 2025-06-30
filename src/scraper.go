package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
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

func fetchHTML(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func extractByXPath(doc *html.Node, xpath string) string {
	if xpath == "" {
		return ""
	}
	
	nodes := htmlquery.Find(doc, xpath)
	if len(nodes) == 0 {
		return ""
	}

	return strings.TrimSpace(htmlquery.InnerText(nodes[0]))
}

func scrapeData(url string, config *XPathConfig) (*ScrapedData, error) {
	doc, err := fetchHTML(url)
	if err != nil {
		return nil, err
	}

	data := &ScrapedData{
		Name:           extractByXPath(doc, config.Name),
		Price:          extractByXPath(doc, config.Price),
		Area:           extractByXPath(doc, config.Area),
		Access:         extractByXPath(doc, config.Access),
		Address:        extractByXPath(doc, config.Address),
		City:           extractByXPath(doc, config.City),
		Prefecture:     extractByXPath(doc, config.Prefecture),
		Contract:       extractByXPath(doc, config.Contract),
		Dept:           extractByXPath(doc, config.Dept),
		Detail:         extractByXPath(doc, config.Detail),
		FacilityName:   extractByXPath(doc, config.FacilityName),
		FacilityType:   extractByXPath(doc, config.FacilityType),
		Holiday:        extractByXPath(doc, config.Holiday),
		License:        extractByXPath(doc, config.License),
		Occupation:     extractByXPath(doc, config.Occupation),
		Position:       extractByXPath(doc, config.Position),
		RequiredSkill:  extractByXPath(doc, config.RequiredSkill),
		StaffComment:   extractByXPath(doc, config.StaffComment),
		Station:        extractByXPath(doc, config.Station),
		WelfareProgram: extractByXPath(doc, config.WelfareProgram),
		WorkingHours:   extractByXPath(doc, config.WorkingHours),
		WorkingStyle:   extractByXPath(doc, config.WorkingStyle),
		TitleOriginal:  extractByXPath(doc, config.TitleOriginal),
	}

	return data, nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run scraper.go <url> <xpath_config.json> <output.json>")
		fmt.Println("Example: go run scraper.go https://example.com/job xpath_config.json result.json")
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
	fmt.Printf("Scraping data from: %s\n", url)
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