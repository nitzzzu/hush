package scanner

import (
"bufio"
"fmt"
"os"
"os/exec"
"path/filepath"
"strings"
)

type Result struct {
File    string
Line    int
Content string
Secret  string
}

type Scanner struct {
secrets map[string]string
}

func New(secrets map[string]string) *Scanner {
return &Scanner{secrets: secrets}
}

func (s *Scanner) ScanFile(path string) ([]Result, error) {
f, err := os.Open(path)
if err != nil {
return nil, err
}
defer f.Close()
var results []Result
sc := bufio.NewScanner(f)
lineNum := 0
for sc.Scan() {
lineNum++
line := sc.Text()
for name, value := range s.secrets {
if value != "" && strings.Contains(line, value) {
results = append(results, Result{File: path, Line: lineNum, Content: line, Secret: name})
break
}
}
}
return results, sc.Err()
}

func (s *Scanner) ScanGitTracked() ([]Result, error) {
files, err := gitTrackedFiles()
if err != nil {
return nil, err
}
return s.scanFiles(files)
}

func (s *Scanner) ScanGitStaged() ([]Result, error) {
files, err := gitStagedFiles()
if err != nil {
return nil, err
}
return s.scanFiles(files)
}

func (s *Scanner) ScanDir(dir string) ([]Result, error) {
var files []string
err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
if err != nil {
return err
}
if !d.IsDir() {
files = append(files, path)
}
return nil
})
if err != nil {
return nil, err
}
return s.scanFiles(files)
}

func (s *Scanner) scanFiles(files []string) ([]Result, error) {
var all []Result
for _, f := range files {
results, err := s.ScanFile(f)
if err != nil {
continue
}
all = append(all, results...)
}
return all, nil
}

func gitTrackedFiles() ([]string, error) {
out, err := exec.Command("git", "ls-files").Output()
if err != nil {
return nil, fmt.Errorf("git ls-files failed: %w", err)
}
return parseFileList(string(out)), nil
}

func gitStagedFiles() ([]string, error) {
out, err := exec.Command("git", "diff", "--cached", "--name-only").Output()
if err != nil {
return nil, fmt.Errorf("git diff --cached failed: %w", err)
}
return parseFileList(string(out)), nil
}

func parseFileList(output string) []string {
var files []string
for _, line := range strings.Split(output, "\n") {
line = strings.TrimSpace(line)
if line != "" {
files = append(files, line)
}
}
return files
}
