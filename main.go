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
	red    = "\033[1;31m%s\033[0m"
	yellow = "\033[1;33m%s\033[0m"
	blue   = "\033[1;34m%s\033[0m"
	purple = "\033[1;35m%s\033[0m"
	cyan   = "\033[1;36m%s\033[0m"
)

var (
	hasGit   = false
	gitColor = ""
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

		checkGit()
		printPrompt(hostname, username, path)

		input, err := reader.ReadString('\n')

		if input == "\n" {
			continue
		}

		if len(removeSpacebar(strings.Split(strings.TrimSuffix(input, "\n"), " "))) == 0 {
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
	args = removeSpacebar(args)

	if len(args) == 0 {
		return nil
	}

	for _, v := range args {
		if v == "<" || v == "<<" || v == ">" || v == ">>" || v == "|" {
			return redirect(input)
		}
	}

	switch args[0] {
	case "cd":
		return cdShell(args)
	case "history":
		return historyShell(args)
	case "vi", "vim":
		return textEditor(args)
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

	if len(args) < 2 || args[1] == "" || args[1] == "~" {
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
	if hasGit == true {

		branch := getGitBranch()

		fmt.Printf(" ")
		fmt.Printf(gitColor, "|")
		fmt.Printf(gitColor, "Git: ")
		fmt.Printf(gitColor, branch)
		fmt.Printf(gitColor, "|")
		fmt.Printf(" ")
	}
	fmt.Printf(cyan, path)
	fmt.Printf("$ ")
}

func removeElemFromSlice(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}

func removeSpacebar(args []string) []string {
	for i := 0; i < len(args); i++ {
		if args[i] == "" {
			args = removeElemFromSlice(args, i)
			i--
		}
	}

	return args
}

func getGitBranch() string {
	cmd := exec.Command("cat", ".git/HEAD")

	stdout, err := cmd.Output()
	if err != nil {
		return ""
	}

	output := strings.TrimSuffix(string(stdout), "\n")
	refs := strings.Split(output, " ")
	branch := strings.Split(refs[1], "/")

	return string(branch[len(branch)-1])
}

func checkGit() {
	if _, err := os.Stat(".git"); !os.IsNotExist(err) {
		hasGit = true

		cmd := exec.Command("bash", "-c", "git status | grep modified")
		stdout, err := cmd.Output()
		if err != nil {
			return
		}

		if len(string(stdout)) != 0 {
			gitColor = red
		} else {
			gitColor = yellow
		}
	} else {
		hasGit = false
	}
}

func redirect(input string) error {
	cmd := exec.Command("bash", "-c", input)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func textEditor(args []string) error {
	if len(args) > 2 {
		return errors.New("ohmyshell: " + args[0] + ": " + "too many arguments")
	}

	if len(args) < 2 {
		cmd := exec.Command(args[0], "--help")

		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		return cmd.Run()
	}

	cmd := exec.Command(args[0], args[1])

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
