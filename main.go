package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

func main() {
	commitsFlag := flag.Bool("commits", false, "Compare with commits")
	tagsFlag := flag.Bool("tags", false, "Compare with tags")
	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Println("Usage: gitmatch [--commits] [--tags] <repository-url>#<branch> <local-directory>")
		fmt.Println("If no flags are specified, defaults to both commits and tags.")
		os.Exit(1)
	}

	split := strings.Split(flag.Args()[0], "#")
	repoURL := split[0]
	var branch string
	if len(split) > 1 {
		branch = split[1]
	}

	localDir := flag.Args()[1]
	repoName := filepath.Base(repoURL)
	repoName = strings.TrimSuffix(repoName, ".git")

	tmpDir := filepath.Join("/tmp", repoName)
	os.RemoveAll(tmpDir)

	var cmd *exec.Cmd
	if branch != "" {
		cmd = exec.Command("git", "clone", "-b", branch, repoURL, tmpDir)
	} else {
		cmd = exec.Command("git", "clone", repoURL, tmpDir)
	}
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error cloning repo: ", err)
		os.Exit(1)
	}

	localHash, err := hashDir(localDir)
	if err != nil {
		fmt.Println("Error hashing local directory: ", err)
		os.Exit(1)
	}

	compareWithCommits := *commitsFlag || !*tagsFlag
	compareWithTags := *tagsFlag || !*commitsFlag

	if compareWithCommits {
		checkCommits(tmpDir, localHash)
	}
	if compareWithTags {
		checkTags(tmpDir, localHash)
	}
}

func checkCommits(tmpDir string, localHash string) {
	cmd := exec.Command("git", "log", "--pretty=format:%H")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting commit hashes: ", err)
		os.Exit(1)
	}
	commits := strings.Split(string(output), "\n")
	for _, commit := range commits {
		cmd = exec.Command("git", "checkout", commit)
		cmd.Dir = tmpDir
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error checking out commit: ", err)
			continue
		}
		commitHash, err := hashDir(tmpDir)
		if err != nil {
			fmt.Println("Error hashing commit: ", err)
			continue
		}
		if commitHash == localHash {
			cmd = exec.Command("git", "show", "-s", "--format=%ci", commit)
			cmd.Dir = tmpDir
			date, err := cmd.Output()
			if err != nil {
				fmt.Println("Error getting commit date: ", err)
			}
			fmt.Println("Found matching commit: ", commit, " Date: ", string(date))
			return
		}
	}
}

func checkTags(tmpDir string, localHash string) {
	cmd := exec.Command("git", "tag", "--list")
	cmd.Dir = tmpDir
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting tags: ", err)
		os.Exit(1)
	}
	tags := strings.Split(string(output), "\n")
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		cmd = exec.Command("git", "checkout", tag)
		cmd.Dir = tmpDir
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error checking out tag: ", err)
			continue
		}
		tagHash, err := hashDir(tmpDir)
		if err != nil {
			fmt.Println("Error hashing tag: ", err)
			continue
		}
		if tagHash == localHash {
			cmd = exec.Command("git", "show", "-s", "--format=%ci", tag)
			cmd.Dir = tmpDir
			date, err := cmd.Output()
			if err != nil {
				fmt.Println("Error getting tag date: ", err)
			}
			fmt.Println("Found matching tag: ", tag, " Date: ", string(date))
			return
		}
	}
}

func hashDir(dir string) (string, error) {
	h := sha256.New()

	var ignore *gitignore.GitIgnore
	gitignorePath := filepath.Join(dir, ".gitignore")

	if _, err := os.Stat(gitignorePath); err == nil {
		ignore, err = gitignore.CompileIgnoreFile(filepath.Join(dir, ".gitignore"))
		if err != nil {
			fmt.Println("Failed to parse .gitignore: ", err)
			ignore = &gitignore.GitIgnore{}
		}
	} else {
		ignore = &gitignore.GitIgnore{}
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
		}

		relativePath, _ := filepath.Rel(dir, path)
		if ignore.MatchesPath(relativePath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(h, file); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
