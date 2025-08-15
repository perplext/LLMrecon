package version

import (
	"strings"
)

// Additional utility functions for the version package can be added here
// Note: ComputeHash and ReadFileContent are already defined in diff.go

// matchPattern checks if a path matches a glob pattern
func matchPattern(path, pattern string) bool {
	matched, err := filepath.Match(pattern, filepath.Base(path))
	if err != nil {
		return false
	}
	
	return matched

// GetFileExtension gets the extension of a file
func GetFileExtension(path string) string {
	return strings.ToLower(filepath.Ext(path))

// IsTextFile checks if a file is a text file based on its extension
func IsTextFile(path string) bool {
	ext := GetFileExtension(path)
	
	textExtensions := map[string]bool{
		".txt":  true,
		".md":   true,
		".json": true,
		".yaml": true,
		".yml":  true,
		".go":   true,
		".py":   true,
		".js":   true,
		".ts":   true,
		".html": true,
		".css":  true,
		".xml":  true,
		".csv":  true,
		".ini":  true,
		".toml": true,
		".sh":   true,
		".bat":  true,
		".ps1":  true,
	}
	
	return textExtensions[ext]

// IsBinaryFile checks if a file is a binary file based on its extension
func IsBinaryFile(path string) bool {
	ext := GetFileExtension(path)
	
	binaryExtensions := map[string]bool{
		".exe":  true,
		".dll":  true,
		".so":   true,
		".dylib": true,
		".bin":  true,
		".obj":  true,
		".o":    true,
		".a":    true,
		".lib":  true,
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".bmp":  true,
		".ico":  true,
		".zip":  true,
		".tar":  true,
		".gz":   true,
		".rar":  true,
		".7z":   true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".ppt":  true,
		".pptx": true,
	}
	
	return binaryExtensions[ext]

// IsCodeFile checks if a file is a code file based on its extension
func IsCodeFile(path string) bool {
	ext := GetFileExtension(path)
	
	codeExtensions := map[string]bool{
		".go":   true,
		".py":   true,
		".js":   true,
		".ts":   true,
		".jsx":  true,
		".tsx":  true,
		".java": true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".hpp":  true,
		".cs":   true,
		".rb":   true,
		".php":  true,
		".swift": true,
		".kt":   true,
		".rs":   true,
		".scala": true,
		".pl":   true,
		".sh":   true,
		".ps1":  true,
		".bat":  true,
		".html": true,
		".css":  true,
		".scss": true,
		".less": true,
		".sql":  true,
	}
	
	return codeExtensions[ext]

// IsConfigFile checks if a file is a configuration file based on its extension
func IsConfigFile(path string) bool {
	ext := GetFileExtension(path)
	basename := filepath.Base(path)
	
	configExtensions := map[string]bool{
		".json": true,
		".yaml": true,
		".yml":  true,
		".toml": true,
		".ini":  true,
		".xml":  true,
		".conf": true,
		".cfg":  true,
		".env":  true,
	}
	
	configFiles := map[string]bool{
		".gitignore":      true,
		".dockerignore":   true,
		"Dockerfile":      true,
		"Makefile":        true,
		"package.json":    true,
		"go.mod":          true,
		"requirements.txt": true,
	}
	
