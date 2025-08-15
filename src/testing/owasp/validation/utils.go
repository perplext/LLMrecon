// Package validation provides methods to validate and detect OWASP LLM vulnerabilities
package validation

import "math"

// MaxInt returns the maximum of two integers
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b

// MinInt returns the minimum of two integers
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b

// MinFloat64 returns the minimum of two float64 values
func MinFloat64(a, b float64) float64 {
	return math.Min(a, b)

// MaxFloat64 returns the maximum of two float64 values
func MaxFloat64(a, b float64) float64 {
	return math.Max(a, b)
