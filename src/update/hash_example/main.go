package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/perplext/LLMrecon/src/update"
)

func main() {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "hash-example")
	if err != nil {
		fmt.Printf("Error creating temp directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFilePath := filepath.Join(tempDir, "test-file.txt")
	testData := []byte("This is test data for hash verification utilities")
	err = os.WriteFile(testFilePath, testData, 0644)
	if err != nil {
		fmt.Printf("Error creating test file: %v\n", err)
		return
	}

	fmt.Println("=== Hash Verification Utilities Example ===")
	fmt.Printf("Test file created at: %s\n\n", testFilePath)

	// Create a hash generator for each algorithm
	algorithms := []update.HashAlgorithm{
		update.SHA256,
		update.SHA512,
		update.SHA1,
		update.MD5,
		update.BLAKE2b,
		update.BLAKE2s,
	}

	fmt.Println("1. Generating hashes for each algorithm:")
	for _, alg := range algorithms {
		generator, err := update.NewHashGenerator(alg)
		if err != nil {
			fmt.Printf("  Error creating %s generator: %v\n", alg, err)
			continue
		}

		result, err := generator.HashFile(testFilePath)
		if err != nil {
			fmt.Printf("  Error generating %s hash: %v\n", alg, err)
			continue
		}

		fmt.Printf("  %s: %s\n", alg, result.HexString)
	}
	fmt.Println()

	// Create a hash verifier
	verifier, err := update.NewHashVerifier()
	if err != nil {
		fmt.Printf("Error creating hash verifier: %v\n", err)
		return
	}

	fmt.Println("2. Generating all hashes at once:")
	allHashes, err := verifier.GenerateFileHashes(testFilePath)
	if err != nil {
		fmt.Printf("  Error generating all hashes: %v\n", err)
	} else {
		for alg, result := range allHashes {
			fmt.Printf("  %s: %s\n", alg, result.HexString)
		}
	}
	fmt.Println()

	// Verify a hash
	fmt.Println("3. Verifying file hash:")
	sha256Hash := allHashes[update.SHA256].String()
	valid, err := verifier.VerifyFileHash(testFilePath, sha256Hash)
	if err != nil {
		fmt.Printf("  Error verifying hash: %v\n", err)
	} else {
		fmt.Printf("  Hash verification result (valid hash): %v\n", valid)
	}

	// Verify an invalid hash
	valid, err = verifier.VerifyFileHash(testFilePath, "sha256:invalidhash")
	if err != nil {
		fmt.Printf("  Error verifying invalid hash: %v\n", err)
	} else {
		fmt.Printf("  Hash verification result (invalid hash): %v\n", valid)
	}
	fmt.Println()

	// Hash data
	fmt.Println("4. Hashing and verifying data:")
	dataToHash := []byte("This is some data to hash")
	dataHash, err := update.HashData(dataToHash, update.SHA256)
	if err != nil {
		fmt.Printf("  Error hashing data: %v\n", err)
	} else {
		fmt.Printf("  SHA-256 hash of data: %s\n", dataHash)

		// Verify data hash
		valid, err := update.VerifyDataHash(dataToHash, dataHash)
		if err != nil {
			fmt.Printf("  Error verifying data hash: %v\n", err)
		} else {
			fmt.Printf("  Data hash verification result: %v\n", valid)
		}
	}
	fmt.Println()

	// Demonstrate secure comparison
	fmt.Println("5. Secure string comparison:")
	fmt.Printf("  'same' == 'same': %v\n", update.SecureCompare("same", "same"))
	fmt.Printf("  'same' == 'different': %v\n", update.SecureCompare("same", "different"))
	fmt.Printf("  'almost same' == 'almost dame': %v\n", update.SecureCompare("almost same", "almost dame"))
	fmt.Println()

	fmt.Println("=== Example Complete ===")
}
