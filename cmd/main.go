package main

import (
	"fmt"
	"os"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error: ", r)
			os.Exit(1)
		}
	}()

}
