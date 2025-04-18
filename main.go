package main

import (
	"fmt"

	"github.com/Ciobi0212/gator.git/internal/config"
)

func main() {
	cfg, err := config.ReadConfig()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Config loaded")

	err = cfg.SetUser("ciobi")

	if err != nil {
		fmt.Println(err)
		return
	}

	cfg, err = config.ReadConfig()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("url: %s\nusername: %s\n", cfg.Db_url, cfg.Current_username)

}
