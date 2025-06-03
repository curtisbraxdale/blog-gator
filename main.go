package main

import (
	"fmt"

	"github.com/curtisbraxdale/blog-gator/internal/config"
)

func main() {
	fig, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config file.")
		return
	}

	err = fig.SetUser("curtisbraxdale")
	if err != nil {
		fmt.Println("Error seting username.")
		return
	}

	fig, err = config.Read()
	if err != nil {
		fmt.Println("Error reading config file.")
		return
	}
	// Print Config Struct To Console
	fmt.Print(fig)
}
