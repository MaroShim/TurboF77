package compiler

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchMatch represents a single search match in project
type SearchMatch struct {
	File    string
	Line    int
	Column  int
	Snippet string
}

// GetSearchRootDir determines the root directory to search for Fortran project
func GetSearchRootDir(targetPath string) string {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		absTarget = targetPath
	}
	fi, err := os.Stat(absTarget)
	if err == nil && !fi.IsDir() {
		return filepath.Dir(absTarget)
	}
	return absTarget
}

// CollectFortranFiles gathers all Fortran source files within rootDir (skipping .git, bin, hidden dirs)
func CollectFortranFiles(rootDir string) []string {
	var files []string
	_ = filepath.Walk(rootDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			base := fi.Name()
			if strings.HasPrefix(base, ".") || base == "bin" || base == "build" || base == "target" {
				return filepath.SkipDir
			}
			return nil
		}
		if IsFortranSource(path) || strings.HasSuffix(strings.ToLower(path), ".f90") || strings.HasSuffix(strings.ToLower(path), ".f95") {
			files = append(files, path)
		}
		return nil
	})
	return files
}

// FindDefinitionInProject searches all Fortran files in the project for the definition of symbol
func FindDefinitionInProject(currentFilePath, symbol string) (defFile string, defLine int, defCol int, found bool) {
	if strings.TrimSpace(symbol) == "" {
		return "", 0, 0, false
	}

	rootDir := GetSearchRootDir(currentFilePath)
	files := CollectFortranFiles(rootDir)

	// Sort files so the current file is checked first
	if currentFilePath != "" {
		absCurrent, _ := filepath.Abs(currentFilePath)
		for i, f := range files {
			absF, _ := filepath.Abs(f)
			if absF == absCurrent {
				if i > 0 {
					files = append([]string{f}, append(files[:i], files[i+1:]...)...)
				}
				break
			}
		}
	}

	escaped := regexp.QuoteMeta(symbol)
	// Fortran subroutine, function, program, entry, module, interface, type:
	// Example:
	//   SUBROUTINE PRINTSUM(...)
	//   INTEGER FUNCTION CALCSUM(...)
	//   MODULE MATH_MOD
	//   TYPE MY_TYPE
	defPattern := regexp.MustCompile(`(?i)^\s*(?:\d+\s+)?(?:(?:integer|real(?:\*[0-9]+)?|double\s+precision|logical|character(?:\*\(?[0-9*]+\)?)?|complex)\s+)?(?:subroutine|function|program|entry|block\s+data|module|interface|type(?:\s*,\s*[^:]+::|\s+))\s*` + escaped + `\b`)

	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Check for Fortran comment lines (C, c, *, ! in col 1 or trimmed !)
			trimmed := strings.TrimSpace(line)
			if len(line) > 0 {
				firstChar := line[0]
				if firstChar == 'C' || firstChar == 'c' || firstChar == '*' || firstChar == '!' {
					continue
				}
			}
			if strings.HasPrefix(trimmed, "!") {
				continue
			}

			if defPattern.MatchString(line) {
				file.Close()
				col := strings.Index(strings.ToLower(line), strings.ToLower(symbol)) + 1
				if col <= 0 {
					col = 1
				}
				return f, lineNum, col, true
			}
		}
		file.Close()
	}

	return "", 0, 0, false
}

// SearchInProject searches for query in all Fortran files in the project
func SearchInProject(currentFilePath, query string, caseSensitive bool) []SearchMatch {
	if strings.TrimSpace(query) == "" {
		return nil
	}

	rootDir := GetSearchRootDir(currentFilePath)
	files := CollectFortranFiles(rootDir)

	var matches []SearchMatch
	targetQuery := query
	if !caseSensitive {
		targetQuery = strings.ToLower(query)
	}

	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			searchIn := line
			if !caseSensitive {
				searchIn = strings.ToLower(line)
			}
			idx := strings.Index(searchIn, targetQuery)
			if idx >= 0 {
				col := idx + 1
				snippet := strings.TrimSpace(line)
				if len(snippet) > 60 {
					snippet = snippet[:60] + "..."
				}
				matches = append(matches, SearchMatch{
					File:    f,
					Line:    lineNum,
					Column:  col,
					Snippet: snippet,
				})
			}
		}
		file.Close()
	}

	return matches
}
