package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/manifoldco/promptui"
)

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
