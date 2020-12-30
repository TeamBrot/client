package main

import (
	"fmt"
)

func main() {
	config, err := GetConfig()
	if err != nil {
		fmt.Println("could not get configuration:", err)
		return
	}
	RunClient(config)
}
