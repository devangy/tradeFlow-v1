package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

func main() {

	// env init
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// proxy server
	username := os.Getenv("username")
	password := os.Getenv("password")
	country := os.Getenv("country")
	entryPoint := os.Getenv("entryPoint")

	// Markets
	kalshi_events_API := os.Getenv("kalshi_events_API")
	poly_events_API := os.Getenv("poly_events_API")

	proxy, err := url.Parse(fmt.Sprintf("http://user-%s-country-%s:%s@%s", username, country, password, entryPoint))
	if err != nil {
		log.Fatalf("Error proxy parsing %v", err)
	}

	// creating a struct instance using a struct literal in memory
	apiClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}

	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	// channel for receiving data
	json_chan := make(chan any, 200)

	go Bot(json_chan)

	go func() {

		// opening a file with append mode for writing data continuously
		// flags append at the end, create if file dont exist and write only to file
		// 0644 unix mode file permision read write execute
		// dir, err := os.Getwd()
		dir, err := os.Getwd()

		if err != nil {
			panic(err)
		}
		file, err := os.OpenFile(filepath.Join(dir, "output.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalln("Err opening file:", err)
		}

		defer file.Close()

		// json encoder for writing directly to file
		encoder := json.NewEncoder(file)
		// encoder.SetIndent("", " ") // make json pretty

		// handler := slog.NewTextHandler(os.Stdout, nil)
		// logger := slog.New(handler)
		// slog.SetDefault(logger)

		for chdata := range json_chan {
			if err = encoder.Encode(chdata); err != nil {
				log.Fatalln("err encoding data to file:", err)
			}
			// slog.Info("LOG", chdata)

		}
		fmt.Println("Finished writing data to output file")
	}()

	for range ticker.C {
		go kalshi(kalshi_events_API, apiClient, json_chan)
		go poly(poly_events_API, apiClient, json_chan)
	}

}

func kalshi(events_API string, apiClient *http.Client, json_chan chan any) {
	// new request

	req, err := http.NewRequest("GET", events_API, nil)
	if err != nil {
		log.Fatalf("Err making get req: %v", err)
	}

	// query params
	//
	params := req.URL.Query()
	params.Add("limit", "5")
	params.Add("status", "open")
	params.Add("with_nested_markets", "true")

	req.URL.RawQuery = params.Encode() // form full URL to make call\

	fmt.Println(req.URL.Query())

	res, err := apiClient.Do(req)
	fmt.Println("url", res.Request.URL)
	if err != nil {
		log.Fatalf("Err getting res: %v ", err)
	}

	defer res.Body.Close() // close connection before exiting
	// read from the Body

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Error reading from res Body")
	}

	type Market struct {
		// OpenInterest int `json:"open_interest"`
		Liquidity       int    `json:"liquidity"`
		Volume          int    `json:"volume"`
		No_ask_dollars  int    `json:"no_ask"`
		Yes_ask_dollars int    `json:"yes_yes"`
		Status          string `json:"status"`
	}

	type Event struct {
		Title        string   `json:"title"`
		EventTicker  string   `json:"event_ticker"`
		SeriesTicker string   `json:"series_ticker"`
		Category     string   `json:"category"`
		Markets      []Market `json:"markets"`
	}

	// initial data struct
	type kmarketdata struct {
		Events []Event
		Cursor string `json:"cursor"`
	}
	var kdata kmarketdata

	// json.NewDecoder(res.Body).Decode(&kdata)

	// unmarshall
	if err = json.Unmarshal(body, &kdata); err != nil {
		log.Fatalf("Error unmarshalling: %v", err)
	}
	fmt.Println("RECEIVED CURSOR:", kdata.Cursor)
	cursor := ""
	cursor = kdata.Cursor
	params.Add("cursor", cursor)

	json_chan <- kdata
}

func poly(events_api string, apiClient *http.Client, json_chan chan any) {
	req, err := http.NewRequest("GET", events_api, nil)
	if err != nil {
		log.Fatal("err making poly GET req", err)
	}

	// structs
	type polymarketdata struct {
		Title    string  `json:"title"`
		Category string  `json:"category"`
		Volume   float64 `json:"volume"`
		Image    string  `json:"image"`
	}

	params := req.URL.Query()

	params.Add("closed", "false")

	res, err := apiClient.Do(req)
	if err != nil {
		log.Fatal("err getting a res", err)
	}

	// creating a new decoder for incmin json data stream
	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		log.Fatal("err reading body", err)
	}

	// decoder := json.NewDecoder(req.Body)

	var pdata []polymarketdata

	// err = json.NewDecoder(res.Body).Decode(&pdata)

	// if err := decoder.Decode(&pdata); err != nil {
	// 	if err == io.EOF {
	// 		return
	// 	}
	// 	log.Println("decode err", err)
	// 	return
	// }

	if err = json.Unmarshal(body, &pdata); err != nil {
		log.Fatal("err unmarshal poly:", err)
	}
	// prettyjson, err := json.MarshalIndent(pdata, " ", "  ")
	// if err != nil {
	// 	log.Fatal("err marshalIndent:", err)
	// }
	// log.Println("poly", string(prettyjson))
	// fmt.Println("Received", string(prettyjson))
	// fmt.Println("fetched at:", time.Now())

	time.Sleep(time.Second)

	json_chan <- pdata

}
