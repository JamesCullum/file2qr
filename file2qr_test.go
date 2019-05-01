package file2qr

import (
	"fmt"
	"os"

	"io"
	"io/ioutil"

	"crypto/sha256"
	"encoding/hex"
)

var osSep string
var testFile string
var testDataRoot string

func init() {
	osSep = string(os.PathSeparator)

	rootDir, _ := os.Getwd()
	testDataRoot = rootDir + osSep + "testdata"

	testFile = testDataRoot + osSep + "1mb.txt"
}

func ExampleNew() {
	converter := New()
	fmt.Println("Default threads:", converter.NumThread)

	converter.NumThread = 10
	fmt.Println("Modified threads:", converter.NumThread)

	// Output:
	// Default threads: 25
	// Modified threads: 10
}

func ExampleConverter_Encode() {
	converter := New()
	progress := 0
	tempFolder := testDataRoot + osSep + "ExampleEncode"

	err := converter.Encode(testFile, tempFolder, &progress)
	files, err2 := ioutil.ReadDir(tempFolder)
	os.RemoveAll(tempFolder)

	fmt.Println("Encoded with errors", err, err2, "into", len(files), "frames with final progress", progress)

	//Output:
	// Encoded with errors <nil> <nil> into 382 frames with final progress 100
}

func ExampleConverter_Decode() {
	converter := New()
	progress := 0
	tempFolder := testDataRoot + osSep + "ExampleDecode"
	tempOutput := testDataRoot + osSep + "ExampleDecode.txt"

	// First encode as in the example for Encode
	converter.Encode(testFile, tempFolder, &progress)

	// Then decode
	progress = 0
	err := converter.Decode(tempFolder, tempOutput, &progress)
	os.RemoveAll(tempFolder)

	// Hash file
	hasher := sha256.New()
	f, err2 := os.Open(tempOutput)
	_, err3 := io.Copy(hasher, f)
	f.Close()
	os.RemoveAll(tempOutput)

	fmt.Println("Decoded with errors", err, err2, err3, "into file with checksum", hex.EncodeToString(hasher.Sum(nil)), "and final progress", progress)

	//Output:
	// Decoded with errors <nil> <nil> <nil> into file with checksum 603fc453b56a70b93befaaeca656a2045968d9ee74cdc7a98f3b41e3b3017169 and final progress 100
}
