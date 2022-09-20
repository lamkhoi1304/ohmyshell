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
	if err := checkFolder(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if err := checkFile(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	history, err := os.OpenFile("/tmp/ohmyshell/history", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	defer history.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		hostname := getHostname()
		username := getUsername()
		path := getPath()

		printPrompt(hostname, username, path)

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

	switch args[0] {
	case "cd":
		return cdShell(args)
	case "history":
		return historyShell(args)
	case "exit":
		exitShell()
	default:
		cmd := exec.Command(args[0], args[1:]...)

		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		return cmd.Run()
	}

	return nil
}

func exitShell() {
	os.Exit(0)
}

func cdShell(args []string) error {
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
}

func historyShell(args []string) error {
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

	return nil
}

func checkFolder() error {
	if _, err := os.Stat("/tmp/ohmyshell"); !os.IsNotExist(err) {
		return err
	}

	return os.Mkdir("/tmp/ohmyshell", 0755)
}

func checkFile() error {
	if _, err := os.Stat("/tmp/ohmyshell"); !os.IsNotExist(err) {
		return err
	}

	_, err := os.Create("/tmp/ohmyshell/history")
	return err
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return hostname
}

func getUsername() string {
	username, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return username.Username
}

func getPath() string {
	path, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return homeSign(path, getUsername())
}

func homeSign(path string, username string) string {
	return strings.Replace(path, "/home/"+username, "~", -1)
}

func printPrompt(hostname string, username string, path string) {
	fmt.Printf(purple, username)
	fmt.Printf(purple, "@")
	fmt.Printf(purple, hostname)
	fmt.Printf(":")
	fmt.Printf(cyan, path)
	fmt.Printf("$ ")
}
