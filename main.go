package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

const (
	purple = "\033[1;35m%s\033[0m"
	cyan   = "\033[1;36m%s\033[0m"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	if _, err := os.Stat("/tmp/ohmyshell"); os.IsNotExist(err) {
		err = os.Mkdir("/tmp/ohmyshell", 0755)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	if _, err := os.Stat("/tmp/ohmyshell"); os.IsNotExist(err) {
		_, err := os.Create("/tmp/ohmyshell/history")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	history, err := os.OpenFile("/tmp/ohmyshell/history", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	defer history.Close()

	for {
		hostname, err := os.Hostname()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		currentUser, err := user.Current()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		path, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		path = strings.Replace(path, "/home/"+currentUser.Username, "~", -1)

		fmt.Printf(purple, currentUser.Username)
		fmt.Printf(purple, "@")
		fmt.Printf(purple, hostname)
		fmt.Printf(":")
		fmt.Printf(cyan, path)
		fmt.Printf("$ ")

		input, err := reader.ReadString('\n')
		if input == "\n" {
			continue
		}
		if _, e := history.Write([]byte(input)); e != nil {
			fmt.Fprintln(os.Stderr, e)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if err := execInput(input); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func execInput(input string) error {
	input = strings.TrimSuffix(input, "\n")
	args := strings.Split(input, " ")

	cmd := exec.Command(args[0], args[1:]...)

	switch args[0] {
	case "cd":
		if len(args) > 2 {
			return errors.New("ohmyshell: " + args[0] + ": " + "too many arguments")
		}
		if len(args) < 2 {
			currentUser, err := user.Current()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return os.Chdir(currentUser.HomeDir)
		}
		if args[1] == "~" {
			currentUser, err := user.Current()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			return os.Chdir(currentUser.HomeDir)
		}

		return os.Chdir(args[1])
	case "history":
		if len(args) > 2 {
			return errors.New("ohmyshell: " + args[0] + ": " + "too many arguments")
		}
		if len(args) < 2 {
			cmd := exec.Command("cat", "/tmp/ohmyshell/history")

			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			return cmd.Run()
		}
		if args[1] == "-c" {
			cmd := exec.Command("truncate", "-s", "0", "/tmp/ohmyshell/history")

			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			return cmd.Run()
		}
	case "exit":
		os.Exit(0)
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
