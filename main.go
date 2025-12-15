package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
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

	// channel where both api will send the json
	json_chan := make(chan any, 200)

	// ctx := context.Background()

	// go Bot()

	go processJson(json_chan)

	// go func() {

	// 	// opening a file with append mode for writing data continuously
	// 	// flags append at the end, create if file dont exist and write only to file
	// 	// 0644 unix mode file permision read write execute
	// 	// dir, err := os.Getwd()
	// 	dir, err := os.Getwd()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	file, err := os.OpenFile(filepath.Join(dir, "output.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	// 	if err != nil {
	// 		log.Fatalln("Err opening file:", err)
	// 	}
	// 	defer file.Close()

	// 	// slog.SetDefault(logger)

	// 	// open file again for scannning lines
	// 	scanFile, err := os.Open("output.jsonl")
	// 	if err != nil {
	// 		log.Fatalln("err opening scanFile", err)
	// 	}
	// 	defer scanFile.Close()

	// 	// scanner := bufio.NewScanner(scanFile)

	// 	// map for storing hash of string with boolean
	// 	seenHash := make(map[[32]byte]bool)

	// 	jdecoder := json.NewDecoder(scanFile)

	// 	for {
	// 		err := jdecoder.Decode()
	// 	}

	// 	for scanner.Scan() {
	// 		line := scanner.Text()
	// 		fmt.Println("line", line)
	// 		// hashing each line using sha256
	// 		hashLine := sha256.Sum256([]byte(line))
	// 		// encoding the hash back to string for checking if it exists in our map
	// 		// stringHash := hex.EncodeToString(hashLine[:])

	// 		fmt.Println("hashLine", hashLine)
	// 		// fmt.Println("stringHash", stringHash)
	// 		// check in hashmap if seen Hash before
	// 		seenHash[hashLine] = true
	// 	}

	// 	for chdata := range json_chan {

	// 		jsonBytes, err := json.Marshal(chdata)
	// 		if err != nil {
	// 			panic(1)
	// 		}

	// 		// conv raw bytes to string
	// 		jsonLine := string(jsonBytes)
	// 		// hash each line
	// 		hashBytes := sha256.Sum256([]byte(jsonLine))
	// 		// convert hashed lines back to string
	// 		// stringHash := hex.EncodeToString(hashLine[:])
	// 		// if the hash is in our map print and move to next iteration and check again
	// 		if seenHash[hashBytes] {
	// 			fmt.Println("Duplicate found")
	// 			continue
	// 		}

	// 		// mark hash seen
	// 		seenHash[hashBytes] = true
	// 		// fmt.Println("map:", seenHash[stringHash])

	// 		// json encoder for writing directly to file
	// 		// Write the bytes we already have + a newline
	// 		if _, err := file.Write(jsonBytes); err != nil {
	// 			log.Println("Write error:", err)
	// 		}
	// 		if _, err := file.WriteString("\n"); err != nil {
	// 			log.Println("Write error:", err)
	// 		}
	// 		// slog.Info("LOG", chdata)

	// 	}
	// 	fmt.Println("Finished writing data to output file")
	// }()

	// var wg sync.WaitGroup
	// wg.Add(2)
	go kalshi(kalshi_events_API, apiClient, json_chan)
	go poly(poly_events_API, apiClient, json_chan)
	// wg.Done()
	// block main forever so program doesnt exit
	select {}

}

func kalshi(events_API string, apiClient *http.Client, json_chan chan any) {

	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	cursor := ""

	for range ticker.C {
		// ptr := &cursor

		req, err := http.NewRequest("GET", events_API, nil)
		if err != nil {
			log.Fatalf("Err making get req: %v", err)
		}

		// query params
		params := req.URL.Query()
		params.Add("limit", "200")
		params.Add("status", "open")
		params.Add("with_nested_markets", "true")
		params.Add("cursor", cursor)

		req.URL.RawQuery = params.Encode() // form full URL to make call\

		fmt.Println(req.URL.Query())
		res, err := apiClient.Do(req)
		// fmt.Println("url", res.Request.URL)
		if err != nil {
			log.Printf("Err getting res: %v", err)
			continue // Skip this loop iteration, try again next tick
		}

		defer res.Body.Close() // close connection before exiting
		// read from the Body

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Panic("Error reading from res Body:", err)
		}

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

		// if err := json.NewDecoder(res.Body).Decode(&kdata); err != nil {
		// 	log.Fatalf("Error decoding: %v", err)

		if err = json.Unmarshal(body, &kdata); err != nil {
			log.Fatalf("Error unmarshalling: %v", err)
		}

		// if err = json.Unmarshal(kdata , &kdatamain); err != nil {
		// 	log.Fatalf("err unmarshal")
		// }
		fmt.Println("RECEIVED CURSOR:", kdata.Cursor)
		// fmt.Println("Data:", kdata)

		// *ptr = kdata.Cursor
		params.Set("cursor", kdata.Cursor)
		// fmt.Println("ptr", ptr)
		cursor = kdata.Cursor
		fmt.Println("cursor", cursor)

		for _, event := range kdata.Events {
			event.Name = "kalshi"
			json_chan <- event
		}

	}

}

func poly(events_api string, apiClient *http.Client, json_chan chan any) {

	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	for range ticker.C {
		req, err := http.NewRequest("GET", events_api, nil)
		if err != nil {
			log.Fatal("err making poly GET req", err)
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

		for _, event := range pdata {
			event.Name = "poly"
			json_chan <- event
		}
	}
}

func processJson(json_chan chan any) {
	// get current dir path
	//

	directory, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current directory path", err)
	}

	// create a file or open existing output.jsonl file for writing data
	file, err := os.OpenFile(filepath.Join(directory, "hashes.bin"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Unable opening file for writing", err)
	}

	defer file.Close()

	// empty hashmap for keeping track of seen hashes with struct as it takes 0 bytes for storage and we care about only the key
	seenMap := make(map[uint64]struct{})

	// start processing json
	for jdata := range json_chan {

		log.Print("JSONLoop:", jdata)

		// convert incoming json to bytes
		jsonBytes, err := json.Marshal(jdata)
		if err != nil {
			log.Fatal("failed converting to jsonBytes", err)
		}

		// init fnv-1a hashing state object
		fnvH := fnv.New64a()
		// hash each json data coming in
		fnvH.Write(jsonBytes)
		// output hashes
		jsonHashValue := fnvH.Sum64()

		log.Print("JsonHashValue", jsonHashValue)

		// if hash seen before jump to next item
		if _, exists := seenMap[jsonHashValue]; exists {
			log.Print("duplicate found")
			continue
		}

		// allocate buffer of size 8 bytes
		var buff [8]byte

		// put the hash value in the buffer in BigEndian byte order
		binary.BigEndian.PutUint64(buff[:], jsonHashValue)

		// add hashes to our map to track seen keys
		// The empty struct takes zero bytes of memory. It has no fields, so it holds no data.
		// struct{}{} we care about only if key exists in collection
		// a way of creating a set data type in Go
		seenMap[jsonHashValue] = struct{}{}

		_, err = file.WriteString(string(buff[:]))
		if err != nil {
			log.Fatal("failed to write buffer to file", err)
		}

	}
}
