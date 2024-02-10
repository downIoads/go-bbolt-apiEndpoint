# go-bbolt-apiEndpoint
Simple API endpoint for bbolt databases, takes path to bbolt db and returns content of database as JSON.

## Usage
Just run with "go run ." and then send a POST request via curl: "curl -X POST -H "Content-Type: application/json" -d '{"input":"./myBboltDb.db"}' localhost:8085/bbolt"
