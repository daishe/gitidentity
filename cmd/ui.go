package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	configv2 "github.com/daishe/gitidentity/config/v2"
	"github.com/daishe/gitidentity/internal/identity"
	"github.com/manifoldco/promptui"
)

func selectIdentityPrompt(ctx context.Context, list []*configv2.Identity) (*configv2.Identity, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("no identities configured")
	}

	stringifiedIdentities := identity.IdentitiesAsStrings(list)
	addMetadataToStringifiedIdentity(ctx, stringifiedIdentities)
	idx, err := selectPrompt("Select identity", stringifiedIdentities)
	if err != nil {
		return nil, err
	}
	if idx < 0 || idx >= len(list) {
		return nil, nil
	}
	return list[idx], nil
}

func addMetadataToStringifiedIdentity(ctx context.Context, stringifiedIdentities []string) {
	current, err := identity.CurrentIdentity(ctx, false)
	currentStr := identity.IdentityAsString(current)
	if err != nil {
		currentStr = "" // setting current to empty will effectively result in skipping 'current' metadata tag
	}
	global, err := identity.GlobalIdentity(ctx)
	globalStr := identity.IdentityAsString(global)
	if err != nil {
		globalStr = "" // setting global to empty will effectively result in skipping 'global' metadata tag
	}
	for idx := range stringifiedIdentities {
		if stringifiedIdentities[idx] == currentStr && currentStr == globalStr {
			stringifiedIdentities[idx] += " (current, global)"
		} else if stringifiedIdentities[idx] == currentStr {
			stringifiedIdentities[idx] += " (current)"
		} else if stringifiedIdentities[idx] == globalStr {
			stringifiedIdentities[idx] += " (global)"
		}
	}
}

func selectPrompt(msg string, list []string) (int, error) {
	if runtime.GOOS == "windows" {
		return selectPrompt_windows(msg, list)
	}

	component := promptui.Select{
		Label:             msg,
		Items:             list,
		Size:              20,
		HideHelp:          true,
		StartInSearchMode: true,
		Searcher: func(input string, i int) bool {
			return strings.Contains(list[i], input)
		},
		Keys: &promptui.SelectKeys{
			Prev:     promptui.Key{Code: readline.CharPrev, Display: "↑"},
			Next:     promptui.Key{Code: readline.CharNext, Display: "↓"},
			PageUp:   promptui.Key{Code: readline.CharBackward, Display: "←"},
			PageDown: promptui.Key{Code: readline.CharForward, Display: "→"},
			Search:   promptui.Key{Code: readline.CharCtrlW, Display: "^W"},
		},
	}

	idx, _, err := component.Run()
	if err != nil {
		return 0, err
	}
	return idx, nil
}

func selectPrompt_windows(msg string, list []string) (int, error) {
	for i, li := range list {
		fmt.Printf("  %d: %s\n", i+1, li)
	}
	fmt.Printf("%s: ", msg)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	fmt.Print("\n")
	if err != nil {
		return -1, err
	}
	idx, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || idx < 1 || idx > len(list) {
		return -1, fmt.Errorf("invalid selection")
	}
	return idx - 1, nil
}
