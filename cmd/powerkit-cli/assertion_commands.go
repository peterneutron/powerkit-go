package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func parseAssertionType(val string) (powerkit.AssertionType, bool) {
	switch val {
	case assertionTypeSystem:
		return powerkit.AssertionTypePreventSystemSleep, true
	case assertionTypeDisplay:
		return powerkit.AssertionTypePreventDisplaySleep, true
	default:
		return 0, false
	}
}

func doAssertionCreate(args []string) {
	if len(args) < 1 {
		log.Fatalf("Error: 'assertion create' requires a type ('system' or 'display').")
	}
	at, ok := parseAssertionType(args[0])
	if !ok {
		log.Fatalf("Error: invalid assertion type '%s'. Use 'system' or 'display'.", args[0])
	}
	reason := "powerkit-cli"
	if len(args) > 1 {
		reason = strings.Join(args[1:], " ")
	}
	id, err := powerkit.CreateAssertion(at, reason)
	if err != nil {
		log.Fatalf("Error creating assertion: %v", err)
	}
	fmt.Printf("Created %s sleep assertion with ID %d.\n", args[0], id)
}

func doAssertionRelease(args []string) {
	if len(args) < 1 {
		log.Fatalf("Error: 'assertion release' requires a type ('system' or 'display').")
	}
	at, ok := parseAssertionType(args[0])
	if !ok {
		log.Fatalf("Error: invalid assertion type '%s'. Use 'system' or 'display'.", args[0])
	}
	powerkit.ReleaseAssertion(at)
	fmt.Printf("Released %s sleep assertion.\n", args[0])
}

func doAssertionStatus(args []string) {
	if len(args) < 1 {
		log.Fatalf("Error: 'assertion status' requires a type ('system' or 'display').")
	}
	at, ok := parseAssertionType(args[0])
	if !ok {
		log.Fatalf("Error: invalid assertion type '%s'. Use 'system' or 'display'.", args[0])
	}
	active := powerkit.IsAssertionActive(at)
	if active {
		id, _ := powerkit.GetAssertionID(at)
		fmt.Printf("%s assertion is ACTIVE (ID %d).\n", args[0], id)
	} else {
		fmt.Printf("%s assertion is not active.\n", args[0])
	}
}

func handleAssertionCommand(args []string) {
	if len(args) < 1 {
		log.Fatalf("Error: 'assertion' requires a subcommand ('create', 'release', 'status').")
	}
	sub := args[0]
	rest := []string{}
	if len(args) > 1 {
		rest = args[1:]
	}
	switch sub {
	case assertionCreate:
		doAssertionCreate(rest)
	case assertionRelease:
		doAssertionRelease(rest)
	case assertionStatus:
		doAssertionStatus(rest)
	default:
		fmt.Printf("Error: unknown subcommand '%s' for 'assertion'.\n\n", sub)
		printUsage()
		os.Exit(1)
	}
}
