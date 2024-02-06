/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/sboon-gg/svctl/cmd"
)

func main() {
	err := cmd.ExecuteNoExit()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
