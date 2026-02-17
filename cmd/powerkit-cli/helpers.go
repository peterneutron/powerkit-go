package main

import (
	"log"
	"os/user"
)

func checkRoot() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Fatal: Could not determine current user: %v", err)
	}
	if currentUser.Uid != "0" {
		log.Fatalf("Error: This command requires root privileges to write to the SMC.\nPlease run with 'sudo'.")
	}
}
