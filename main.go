package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

func main() {

	// enable logging debug level
	log.SetLevel(log.DebugLevel)

	fmt.Printf(`
[38;2;120;200;255m888                                    888                  [0m
[38;2;130;190;255m888                                    888                  [0m
[38;2;140;180;255m888                                    888                  [0m
[38;2;150;170;255m888  888 888d888 8888b.  888  888  888 888  .d88b.  888d888 [0m
[38;2;160;160;255m888 .88P 888P"      "88b 888  888  888 888 d8P  Y8b 888P"    [0m
[38;2;170;150;255m888888K  888    .d888888 888  888  888 888 88888888 888     [0m
[38;2;180;140;255m888 "88b 888    888  888 Y88b 888 d88P 888 Y8b.     888     [0m
[38;2;190;130;255m888  888 888    "Y888888  "Y8888888P"  888  "Y8888  888     [0m


`)

	// wait 3 second before starting up
	seconds := 3
	for i := seconds; i > 0; i-- {
		fmt.Printf(
			"\r\033[38;5;82mStarting in... \033[38;5;196m%d\033[0m \n",
			i,
		)

		time.Sleep(time.Second)
	}

	// env init
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Loading .env File")
	}
	// proxy server creds
	// username := os.Getenv("username")
	// password := os.Getenv("password")
	// country := os.Getenv("country")
	// entryPoint := os.Getenv("entryPoint")

	// Markets
	// kalshi_events_API := os.Getenv("kalshi_events_API")
	// poly_events_API := os.Getenv("poly_events_API")
	poly_trades_API := os.Getenv("poly_trades_API")
	poly_profile_API := os.Getenv("poly_walletProfile_API")
	// kalshi_trades_API := os.Getenv("kalshi_trades_API")

	// proxy server provider URL for rotating proxy
	// proxy, err := url.Parse(fmt.Sprintf("http://user-%s-country-%s:%s@%s", username, country, password, entryPoint))
	// if err != nil {
	// 	log.Fatalf("Error proxy parsing %v", err)
	// }

	// creating a struct instance using a struct literal in memory
	apiClient := &http.Client{
		Timeout: 30 * time.Second,
		// Transport: &http.Transport{
		// 	Proxy: http.ProxyURL(proxy),
		// },
	}

	err = os.MkdirAll("logs", 0755)
	if err != nil {
		log.Fatal("Creating logs dir: ", err)
	}

	// channel where both api will send the json
	// events_chan := make(chan any, 200)
	// // Telegram channel for clean and filtered data according to logic applied
	// tgEventC := make(chan any, 200)
	// trade wallet address chan
	tradeWalletC := make(chan Trade, 200)

	// go Bot(tgEventC)

	// go kalshi(kalshi_events_API, apiClient, events_chan)
	// log.Info("Started Kalshi Events Worker")
	// go poly(poly_events_API, apiClient, events_chan)
	// log.Info("Started Poly Events Worker")

	// go processEvents(events_chan, tgEventC)
	// log.Info("Started Events Processing Worker")

	go polyTrades(poly_trades_API, apiClient, tradeWalletC)
	log.Info("Started Poly Trades Worker")

	go polyWallet(poly_profile_API, apiClient, tradeWalletC)
	log.Info("Started Poly Trades Worker")
	// go kalshiTrades(kalshi_trades_API, apiClient)
	// log.Info("Started Kalshi Trades Worker")

	select {}
}

func kalshi(events_API string, apiClient *http.Client, events_chan chan any) {

	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	for range ticker.C {
		cursor := ""

		go func() {
			req, err := http.NewRequest("GET", events_API, nil)
			if err != nil {
				log.Error("GET request failed Kalshi Events", err)
				return
			}

			// query params
			params := req.URL.Query()
			params.Add("limit", "200")
			params.Add("status", "open")
			params.Add("with_nested_markets", "true")
			params.Add("cursor", cursor)

			req.URL.RawQuery = params.Encode() // form full URL to make call\

			log.Debug("Kalshi req query: ", req.URL.Query())
			res, err := apiClient.Do(req)
			if err != nil {
				log.Debug("Kalshi statcode: ", res.StatusCode)
				log.Error("failed response Kalshi Events", err)
				return
			}

			defer res.Body.Close()

			// backoff for 15 second if server sends 2 many reqs status
			var backoff time.Duration

			if res.StatusCode == http.StatusTooManyRequests {
				backoff = 15 * time.Second
				return
			}

			if backoff > 0 {
				time.Sleep(backoff)
				backoff = 0
			}

			// read from the Body
			//
			// log.Debug("kalsi", res.StatusCode)

			// body, err := io.ReadAll(res.Body)
			// if err != nil {
			// 	log.Error("reading from response body Kalshi Events: ", err)
			// }

			// type Market struct {
			// 	// OpenInterest int `json:"open_interest"`
			// 	Liquidity       int    `json:"liquidity"`
			// 	Volume          int    `json:"volume"`
			// 	No_ask_dollars  int    `json:"no_ask"`
			// 	Yes_ask_dollars int    `json:"yes_yes"`
			// 	Status          string `json:"status"`
			// }

			type Event struct {
				Name         string
				Title        string `json:"title"`
				EventTicker  string `json:"event_ticker"`
				SeriesTicker string `json:"series_ticker"`
				Category     string `json:"category"`
			}

			// initial data struct
			type kdump struct {
				Events []Event
				Cursor string `json:"cursor"`
			}
			var kdata kdump

			if err := json.NewDecoder(res.Body).Decode(&kdata); err != nil {
				log.Debug("[KalshiEvents]| decoding Body Kalshi: ", err)
			}

			// if err = json.Unmarshal(body, &kdata); err != nil {
			// 	log.Error("reading from: ", err)
			// }

			log.Debug("KALSHI RECEIVED CURSOR:", kdata.Cursor)
			// fmt.Println("Data:", kdata)

			// *ptr = kdata.Cursor
			params.Set("cursor", kdata.Cursor)
			// fmt.Println("ptr", ptr)
			cursor = kdata.Cursor
			log.Debug("CursorValue: ", cursor)

			for _, event := range kdata.Events {
				event.Name = "kalshi"
				events_chan <- event
			}

		}()

	}

}

func poly(events_api string, apiClient *http.Client, events_chan chan any) {

	ticker := time.NewTicker(150 * time.Millisecond)

	defer ticker.Stop()

	for range ticker.C {

		go func() {
			req, err := http.NewRequest("GET", events_api, nil)
			if err != nil {
				log.Error("[PolyEvents] GET request: ", err)
			}

			// structs

			type polymarketdata struct {
				Name     string
				Title    string  `json:"title"`
				Category string  `json:"category"`
				Volume   float64 `json:"volume"`
				Image    string  `json:"image"`
			}

			params := req.URL.Query()

			params.Add("closed", "false")

			res, err := apiClient.Do(req)
			if err != nil {
				log.Error("[PolyEvents] | getting response", err)
				return
			}

			// creating a new decoder for incmin json data stream
			body, err := io.ReadAll(res.Body)
			defer res.Body.Close()

			if err != nil {
				log.Error("[PolyEvents] | Reading body", err)
			}

			var pdata []polymarketdata

			// decoder := json.NewDecoder(req.Body)

			err = json.NewDecoder(res.Body).Decode(&pdata)

			// if err := decoder.Decode(&pdata); err != nil {
			// 	if err == io.EOF {
			// 		return
			// 	}
			// 	log.Fatal("[PolyEvent] | Decoding response body: ", err)
			// 	return
			// }

			if err = json.Unmarshal(body, &pdata); err != nil {
				log.Error("err unmarshal poly:", err)
			}

			for _, event := range pdata {
				event.Name = "poly"
				events_chan <- event
			}
		}()
	}

}

func processEvents(events_chan chan any, tgEventsC chan any) {
	// get current dir path
	//

	directory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current directory path: ", err)
	}

	logsDir := filepath.Join(directory, "logs")
	log.Info("created logs dir")

	// create a file or open existing output.jsonl file for writing data
	file, err := os.OpenFile(filepath.Join(logsDir, "eventsHash.bin"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Opening hashes file for writing: ", err)
	}

	defer file.Close()

	// empty hashmap for keeping track of seen hashes with struct as it takes 0 bytes for storage and we care about only the key
	seenMap := make(map[uint64]struct{})

	// load already seen hashes into hashmap so that dupes dont get forwarded
	binFile, err := os.Open(filepath.Join(logsDir, "eventsHash.bin"))
	if err != nil {
		log.Fatal("Open hashes bin file: ", err)

	}
	defer binFile.Close()

	// channel for sending only freshdata

	for {

		var buffBin [8]byte

		_, err = io.ReadFull(binFile, buffBin[:])

		if err == io.EOF {
			log.Info("[processEvents] | EOF Reached!")
			break
		}
		if err != nil {
			log.Error("Unexpected error reading bin file:", err)
			break
		}

		var value uint64
		value = binary.BigEndian.Uint64(buffBin[:])
		// load hashes in hashmap
		seenMap[value] = struct{}{}

		log.Debug("value", value)

		log.Debug("buffBin", buffBin)
	}

	// start processing json
	for jdata := range events_chan {

		log.Debug("JSON_chanData: ", jdata)

		// convert incoming json to bytes
		jsonBytes, err := json.Marshal(jdata)
		if err != nil {
			log.Error("[processEvent] | Converting to bytes: ", err)
		}

		// init fnv-1a hashing state object
		fnvH := fnv.New64a()
		// hash each json data coming in
		fnvH.Write(jsonBytes)
		// output hashes
		jsonHashValue := fnvH.Sum64()

		log.Debug("JsonHashValue", jsonHashValue)

		// if hash seen before jump to next item
		if _, exists := seenMap[jsonHashValue]; exists {
			log.Debug("Duplicate found skipping")
			continue
		}

		// send only fresh data to Tg bot
		tgEventsC <- jdata

		// allocate buffer of size 8 bytes
		var buff [8]byte

		log.Debug("empty buff", buff)

		// put the hash value in the buffer in BigEndian byte order
		binary.BigEndian.PutUint64(buff[:], jsonHashValue)

		// add hashes to our map to track seen keys
		// The empty struct takes zero bytes of memory. It has no fields, so it holds no data.
		// struct{}{} we care about only if key exists in collection
		// a way of creating a set data type in Go
		seenMap[jsonHashValue] = struct{}{}

		bywritten, err := file.Write(buff[:])
		log.Debug("bytes written: ", bywritten)
		if err != nil {
			log.Error("[processEvent] | Write buffer to hashes.bin file: ", err)
		}

	}
}

type Trade struct {
	ProxyWallet           string  `json:"proxyWallet"`
	Side                  string  `json:"side"`
	Asset                 string  `json:"asset"`
	ConditionID           string  `json:"conditionId"`
	Size                  float64 `json:"size"`
	Price                 float64 `json:"price"`
	Timestamp             int64   `json:"timestamp"`
	Title                 string  `json:"title"`
	Slug                  string  `json:"slug"`
	Icon                  string  `json:"icon"`
	EventSlug             string  `json:"eventSlug"`
	Outcome               string  `json:"outcome"`
	OutcomeIndex          int     `json:"outcomeIndex"`
	Name                  string  `json:"name"`
	Pseudonym             string  `json:"pseudonym"`
	Bio                   string  `json:"bio"`
	ProfileImage          string  `json:"profileImage"`
	ProfileImageOptimized string  `json:"profileImageOptimized"`
	TransactionHash       string  `json:"transactionHash"`
	TradeSum              float64
}

func polyTrades(api string, apiClient *http.Client, tradeWalletC chan Trade) {
	ticker := time.NewTicker(150 * time.Millisecond)

	directory, err := os.Getwd()
	if err != nil {
		log.Fatal("opening polyTrades file", err)
	}

	logsPath := filepath.Join(directory, "logs")

	ptradesF, _ := os.OpenFile(filepath.Join(logsPath, "polyTrades.jsonl"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	tradesMap := make(map[string]struct{})

	for range ticker.C {
		go func() {
			req, err := http.NewRequest("GET", api, nil)
			if err != nil {
				log.Error("[polyTrades] | GET request: ", err)
			}

			res, err := apiClient.Do(req)
			if err != nil {
				log.Error("[polyTrades] | Failed response: ", err)
				return

			}
			defer res.Body.Close()

			log.Debug("status", res.StatusCode)

			if res.StatusCode == http.StatusTooManyRequests {
				log.Error("[polyTrades] | Status 429: ", err)
				time.Sleep(15 * time.Second)
			}

			// backoff for 15 second if server sends 2 many reqs status
			var backoff time.Duration

			if res.StatusCode == http.StatusTooManyRequests {
				backoff = 15 * time.Second
				return
			}

			if backoff > 0 {
				time.Sleep(backoff)
				backoff = 0
			}

			// type Trade struct {
			// 	ProxyWallet           string  `json:"proxyWallet"`
			// 	Side                  string  `json:"side"`
			// 	Asset                 string  `json:"asset"`
			// 	ConditionID           string  `json:"conditionId"`
			// 	Size                  float64 `json:"size"`
			// 	Price                 float64 `json:"price"`
			// 	Timestamp             int64   `json:"timestamp"`
			// 	Title                 string  `json:"title"`
			// 	Slug                  string  `json:"slug"`
			// 	Icon                  string  `json:"icon"`
			// 	EventSlug             string  `json:"eventSlug"`
			// 	Outcome               string  `json:"outcome"`
			// 	OutcomeIndex          int     `json:"outcomeIndex"`
			// 	Name                  string  `json:"name"`
			// 	Pseudonym             string  `json:"pseudonym"`
			// 	Bio                   string  `json:"bio"`
			// 	ProfileImage          string  `json:"profileImage"`
			// 	ProfileImageOptimized string  `json:"profileImageOptimized"`
			// 	TransactionHash       string  `json:"transactionHash"`
			// 	TradeSum              float64
			// }

			var trades []Trade

			err = json.NewDecoder(res.Body).Decode(&trades)
			if err != nil {
				log.Error("[polyTrades] | Response Body decoding: ", err)
				return
			}

			left := 0
			for right := 0; right < len(trades); right++ {

				const windowMillisec = int64(120 * 1000)

				// window invalidated remove the trades outside our time window
				for left <= right && (trades[right].Timestamp-trades[left].Timestamp) > windowMillisec {
					left++
				}

				// map returns value and boolean for key present or not
				if _, exists := tradesMap[trades[right].TransactionHash]; exists {
					continue
				}

				tradeSum := trades[right].Size * trades[right].Price

				thresholds := []float64{500, 1000, 5000, 10000}

				for _, value := range thresholds {
					if tradeSum >= value {
						log.Debug("Value:", tradeSum)
						tradesMap[trades[right].TransactionHash] = struct{}{}
						trades[right].TradeSum = tradeSum
						tradeWalletC <- trades[right]
						prettyJson, _ := json.MarshalIndent(trades[right], "", "  ")
						log.Debug("LT ❤️:", string(prettyJson))
						ptradesF.WriteString(string(prettyJson) + "\n")
					}
				}
			}
		}()
	}
}

func polyWallet(api string, apiClient *http.Client, tradeWalletC chan Trade) {
	// dir, err := os.Getwd()
	// if err != nil {
	// 	log.Error("polyTrades | Error getting dir path:")
	// 	return
	// }

	// tradesDir := filepath.Join(dir, "logs")

	// tradesF, err := os.OpenFile(filepath.Join(tradesDir, "polyTrades.jsonl"), os.O_RDONLY, 0)
	// if err != nil {
	// 	log.Fatal("Opening hashes file for writing: ", err)
	// }

	type UserTrades struct {
		ProxyWallet     string  `json:"proxyWallet"`
		Asset           string  `json:"asset"`
		ConditionID     string  `json:"conditionId"`
		AvgPrice        float64 `json:"avgPrice"`
		TotalBought     float64 `json:"totalBought"`
		RealizedPnl     float64 `json:"realizedPnl"`
		CurPrice        float64 `json:"curPrice"`
		Timestamp       int64   `json:"timestamp"`
		Title           string  `json:"title"`
		Slug            string  `json:"slug"`
		Icon            string  `json:"icon"`
		EventSlug       string  `json:"eventSlug"`
		Outcome         string  `json:"outcome"`
		OutcomeIndex    int     `json:"outcomeIndex"`
		OppositeOutcome string  `json:"oppositeOutcome"`
		OppositeAsset   string  `json:"oppositeAsset"`
		EndDate         string  `json:"endDate"`
	}

	// type Trade struct {
	// 	ProxyWallet           string  `json:"proxyWallet"`
	// 	Side                  string  `json:"side"`
	// 	Asset                 string  `json:"asset"`
	// 	ConditionID           string  `json:"conditionId"`
	// 	Size                  float64 `json:"size"`
	// 	Price                 float64 `json:"price"`
	// 	Timestamp             int64   `json:"timestamp"`
	// 	Title                 string  `json:"title"`
	// 	Slug                  string  `json:"slug"`
	// 	Icon                  string  `json:"icon"`
	// 	EventSlug             string  `json:"eventSlug"`
	// 	Outcome               string  `json:"outcome"`
	// 	OutcomeIndex          int     `json:"outcomeIndex"`
	// 	Name                  string  `json:"name"`
	// 	Pseudonym             string  `json:"pseudonym"`
	// 	Bio                   string  `json:"bio"`
	// 	ProfileImage          string  `json:"profileImage"`
	// 	ProfileImageOptimized string  `json:"profileImageOptimized"`
	// 	TransactionHash       string  `json:"transactionHash"`
	// 	TradeSum              float64
	// }

	var users []UserTrades

	// var trades Trade

	// incoming trade struct iteration from channel
	for trade := range tradeWalletC {

		log.Print("tradePWallet:", trade)
		// extrac

		go func() {
			req, err := http.NewRequest("GET", api, nil)
			if err != nil {
				log.Error("creating request polyWallet: ", err)
				return
			}

			// params
			params := req.URL.Query()
			// extract wallet addr from each trade
			params.Add("user", trade.ProxyWallet)
			params.Add("limit", strconv.Itoa(100))
			params.Add("sortBy", "TIMESTAMP")
			params.Add("sortDirection", "DESC")

			req.URL.RawQuery = params.Encode()
			log.Print("API:", api)

			res, err := apiClient.Do(req)
			if err != nil {
				log.Error("polyWallet - failed to get response: ", err)
				return
			}

			log.Info("polyWallet: ", res.StatusCode)

			json.NewDecoder(res.Body).Decode(&users)

			for _, userTrade := range users {
				log.Debug("user:", userTrade)

			}

		}()

	}

}

func kalshiTrades(api string, apiClient *http.Client) {
	ticker := time.NewTicker(150 * time.Millisecond)

	defer ticker.Stop()

	for range ticker.C {
		go func() {
			req, err := http.NewRequest("GET", api, nil)
			if err != nil {
				log.Error("[kalshiTrades] | GET request: ", err)
			}

			res, err := apiClient.Do(req)
			if err != nil {
				log.Error("[kalshiTrades] | GET response: ", err)
			}

			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Error("[kalshiTrades] | Parse response body: ", err)
			}

			log.Debug("ktrades: ", string(body))
		}()
	}
}
