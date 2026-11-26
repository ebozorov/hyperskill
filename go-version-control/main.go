package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	vcsDir     = "./vcs"
	configFile = vcsDir + "/config.txt"
	indexFile  = vcsDir + "/index.txt"
	logFile    = vcsDir + "/log.txt"
	commitsDir = vcsDir + "/commits"
)

func main() {
	// Ensure vcs directory exists
	err := os.MkdirAll(vcsDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args[1:]
	if len(args) == 0 || os.Args[1] == "--help" {
		fmt.Println("These are SVCS commands:")
		fmt.Println("config     Get and set a username.")
		fmt.Println("add        Add a file to the index.")
		fmt.Println("log        Show commit logs.")
		fmt.Println("commit     Save changes.")
		fmt.Println("checkout   Restore a file.")
		return
	}

	switch args[0] {
	case "config":
		handleConfig(args[1:])
	case "add":
		handleAdd(args[1:])
	case "log":
		handleLog()
	case "commit":
		handleCommit(args[1:])
	case "checkout":
		handleCheckout(args[1:])
	default:
		fmt.Printf("'%s' is not a SVCS command.\n", os.Args[1])
	}
}

func handleConfig(args []string) {
	if len(args) == 0 {
		// Show existing username
		data, err := os.ReadFile(configFile)
		if err != nil || len(strings.TrimSpace(string(data))) == 0 {
			fmt.Println("Please, tell me who you are.")
			return
		}
		fmt.Printf("The username is %s.\n", strings.TrimSpace(string(data)))
	} else {
		// Set new username
		name := args[0]
		err := os.WriteFile(configFile, []byte(name), 0644)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The username is %s.\n", name)
	}
}

func handleAdd(args []string) {
	if len(args) == 0 {
		// Show tracked files
		data, err := os.ReadFile(indexFile)
		if err != nil || len(strings.TrimSpace(string(data))) == 0 {
			fmt.Println("Add a file to the index.")
			return
		}
		fmt.Println("Tracked files:")
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	} else {
		fileName := args[0]
		// Check if file exists
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			fmt.Printf("Can't find '%s'.\n", fileName)
			return
		}
		// Append to index
		file, err := os.OpenFile(indexFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// Check if already tracked
		data, _ := os.ReadFile(indexFile)
		if strings.Contains(string(data), fileName) {
			fmt.Printf("The file '%s' is tracked.\n", fileName)
			return
		}

		_, err = file.WriteString(fileName + "\n")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The file '%s' is tracked.\n", fileName)
	}
}

func handleCommit(args []string) {
	if len(args) == 0 {
		fmt.Println("Message was not passed.")
		return
	}
	message := strings.Join(args, " ")

	// Read tracked files
	data, err := os.ReadFile(indexFile)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println("Nothing to commit.")
		return
	}
	files := strings.Split(strings.TrimSpace(string(data)), "\n")

	// Compute hash of all tracked files
	h := sha256.New()
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			continue
		}
		io.Copy(h, file)
		file.Close()
	}
	commitID := hex.EncodeToString(h.Sum(nil))[:40]

	// Check if commit already exists
	if _, err := os.Stat(filepath.Join(commitsDir, commitID)); err == nil {
		fmt.Println("Nothing to commit.")
		return
	}

	// Create commit directory
	commitPath := filepath.Join(commitsDir, commitID)
	err = os.MkdirAll(commitPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// Copy tracked files into commit directory
	for _, f := range files {
		src, err := os.Open(f)
		if err != nil {
			continue
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(commitPath, filepath.Base(f)))
		if err != nil {
			log.Fatal(err)
		}
		defer dst.Close()

		io.Copy(dst, src)
	}

	// Get author
	author := "Unknown"
	if name, err := os.ReadFile(configFile); err == nil {
		author = strings.TrimSpace(string(name))
	}

	// Append to log.txt
	entry := fmt.Sprintf("commit %s\nAuthor: %s\n%s\n", commitID, author, message)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(entry + "\n")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Changes are committed.")
}

func handleLog() {
	data, err := os.ReadFile(logFile)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println("No commits yet.")
		return
	}
	entries := strings.Split(strings.TrimSpace(string(data)), "\n\n")
	// Reverse order
	for i := len(entries) - 1; i >= 0; i-- {
		fmt.Println(entries[i])
	}
}

func handleCheckout(args []string) {
	if len(args) == 0 {
		fmt.Println("Commit id was not passed.")
		return
	}
	commitID := args[0]
	commitPath := filepath.Join(commitsDir, commitID)

	// Check if commit exists
	if _, err := os.Stat(commitPath); os.IsNotExist(err) {
		fmt.Println("Commit does not exist.")
		return
	}

	// Read tracked files from index
	data, err := os.ReadFile(indexFile)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		fmt.Println("No tracked files to restore.")
		return
	}
	files := strings.Split(strings.TrimSpace(string(data)), "\n")

	// Restore each tracked file from commit directory
	for _, f := range files {
		srcPath := filepath.Join(commitPath, filepath.Base(f))
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		src, err := os.Open(srcPath)
		if err != nil {
			continue
		}
		defer src.Close()

		dst, err := os.Create(f)
		if err != nil {
			log.Fatal(err)
		}
		defer dst.Close()

		io.Copy(dst, src)
	}

	fmt.Printf("Switched to commit %s.\n", commitID)
}
