package main

import (
	"fmt"

	"github.com/crisp-coder/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("%w", err)
		return
	}
	err = cfg.SetUser("crisp-coder")
	if err != nil {
		fmt.Println("%w", err)
		return
	}
	cfg2, err := config.Read()
	if err != nil {
		fmt.Println("%w", err)
		return
	}
	fmt.Printf("DBCONN: %s\nUser: %s\n", cfg2.Db_url, cfg2.Username)
}
