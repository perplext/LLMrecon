package prompt

// Shared utility functions for the prompt security package

// getContext extracts context around a match with configurable context size
func getContext(text string, start, end int) string {
	return getContextWithSize(text, start, end, 50)
}

// getContextWithSize extracts context around a match with specified context size
func getContextWithSize(text string, start, end int, contextSize int) string {
	contextStart := start - contextSize
	if contextStart < 0 {
		contextStart = 0
	}
	
	contextEnd := end + contextSize
	if contextEnd > len(text) {
		contextEnd = len(text)
	}
	
	return text[contextStart:contextEnd]
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}