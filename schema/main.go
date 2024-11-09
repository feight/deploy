package main

import (
	"encoding/json"
	"os"

	"github.com/feight/deploy/deploy"
)

func main() {
	schema, _ := GetSchema(deploy.Config{})
	b, _ := json.MarshalIndent(schema, "", "    ")
	os.WriteFile("schema.json", b, os.ModePerm)
}
