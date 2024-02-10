package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http" 		// API endpoints

	bolt "go.etcd.io/bbolt"
)


// ---- Bbolt related code ----

// BboltDb is a struct representing a bbolt database.
type BboltDb struct {
	Path string 							`json:"path"`		// path to db file (this data is received from Swift program) 
	Buckets map[string]map[string]string 	`json:"buckets"`	// map each Bucket to the key-value pairs it contains
}

// GetDbContentAsJson takes the path to a bbolt database, reads all its content and returns it as a serialized JSON object of BboltDb along with an error.
func GetDbContentAsJson(dbPath string) ([]byte, error) {
	var bboltDbObject BboltDb

	// intialize the Buckets map
	bboltDbObject.Buckets = make(map[string]map[string]string)

	// open database
	dbInstance, err := bolt.Open(dbPath, 0400, nil) // 0400 == read only
	if err != nil {
		return nil, fmt.Errorf("Failed to open database: %v\n", err)
	}
	defer dbInstance.Close()

	// get existing buckets
	err = dbInstance.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(bucketName []byte, _ *bolt.Bucket) error {
			// create new empty bucket that represents the bucket we just found
			bboltDbObject.Buckets[string(bucketName)] = make(map[string]string)
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to get buckets of database due to error: %v\n", err)
	}

	// iterate over each bucket
	for bucketNameString := range bboltDbObject.Buckets {
		// populate bboltDbObject with data
		err = dbInstance.View(func(tx *bolt.Tx) error {
			// access current bucket
	        b := tx.Bucket([]byte(bucketNameString))
	        if b == nil {
	            return fmt.Errorf("Failed to access bucket %v even though it should exist!\n", bucketNameString)
	        }
	        // iterate over each key in current bucket
	        cursor := b.Cursor()
	        for keyBytes, _ := cursor.First(); keyBytes != nil; keyBytes, _ = cursor.Next() {

				// cast key to string
	        	keyString := hex.EncodeToString(keyBytes)

	        	// get value that corresponds to this key
	        	v := b.Get(keyBytes)
			    if v == nil {
			    	return fmt.Errorf("In bucket %v tried to access value of key %v but failed due to error: %v\n", bucketNameString, keyString, err)
			    }

	        	// add key-value pair to bboltDbObject in the correct bucket
	            bboltDbObject.Buckets[bucketNameString][keyString] = string(v)
	        }

	        return nil
	    })
	    if err != nil {
	        panic(err)
	    }

	}

	// serialize bboltDbObject to json
	bboltDbObjectJson, err := json.Marshal(bboltDbObject)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize object to json: %v\n", err)
	}

	return bboltDbObjectJson, nil
}


// ---- API endpoints related code ----

// RequestPayload is a struct representing the expected request payload
type RequestPayload struct {
	Input string `json:"input"`
}

// ResponsePayload is a struct representing the response payload
type ResponsePayload struct {
	Result string `json:"result"`
}

// handleRequest handles API endpoint requests
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// only allow POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Please use POST.", http.StatusMethodNotAllowed)
		return
	}

	// decode request
	var requestPayload RequestPayload
	err := json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// do actual work
	resultBytes, err := GetDbContentAsJson(requestPayload.Input)
	if err != nil {
		fmt.Println("ERROR:", err)
		return // if the request is valid but the response invalid, then do not respond
	}
	result := string(resultBytes)

	// create response payload
	responsePayload := ResponsePayload {
		Result: result,
	}

	// encode response payload and send it
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(responsePayload)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fmt.Println("Successfully sent response.")
}

func main() {
	API_ENDPOINT := "/bbolt"
	PORT := 8085

	http.HandleFunc(API_ENDPOINT, handleRequest)
	fmt.Println("Server listening on localhost:" + fmt.Sprint(PORT) + API_ENDPOINT)
	http.ListenAndServe(":" + fmt.Sprint(PORT), nil)

	// SEND EXAMPLE REQUEST:
	// 		curl -X POST -H "Content-Type: application/json" -d '{"input":"./myBboltDb.db"}' localhost:8085/bbolt

	// if you put path to non-existing database, response will be: {"result":"{\"path\":\"\",\"buckets\":{}}"}
}
