package main

import (
	"fmt"

	"github.com/atroxxxxxx/embed-store/internal/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic("flag parsing failed: " + err.Error())
	}
	fmt.Println(cfg)
}
