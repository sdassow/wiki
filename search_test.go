package main

import (
	"bufio"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// taken from https://stackoverflow.com/a/18479916/13124
func readLines(path string) ([]string, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	var lines []string
	size := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		size++
	}
	return lines, size, scanner.Err()
}

func generateTestData(folder string) []string {
	dictionaryFile := "H:\\dev.external\\dictionary\\popular.txt" //Taken from https://github.com/dolph/dictionary/blob/master/popular.txt
	dictionary, dicSize, _ := readLines(dictionaryFile)
	usedWords := make([]string, 0)

	numFiles := r.Intn(5) + 1
	for i := 0; i <= numFiles; i++ {
		fileName := dictionary[r.Intn(dicSize)] + ".txt"
		filePath := filepath.Join(folder, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			continue
		}
		writer := bufio.NewWriter(file)

		numSentances := r.Intn(10) + 1
		for n := 0; n <= numSentances; n++ {
			var sentance strings.Builder

			numWords := r.Intn(15) + 1
			for k := 0; k <= numWords; k++ {
				nextWord := dictionary[r.Intn(dicSize)]
				usedWords = append(usedWords, nextWord)

				method := r.Intn(11)
				switch method {
				case 10:
					sentance.WriteString("# ")
					sentance.WriteString(nextWord)
					sentance.WriteString("\r\n\r\n")
					break
				case 9:
					sentance.WriteString("* ")
					sentance.WriteString(nextWord)
					sentance.WriteString("\r\n\r\n")
					break
				case 8, 7, 6:
					sentance.WriteString("[")
					sentance.WriteString(nextWord)
					sentance.WriteString("]")
					sentance.WriteString("(")
					sentance.WriteString(nextWord)
					sentance.WriteString(")")
					sentance.WriteString(" ")
					break
				default:
					sentance.WriteString(nextWord)
					sentance.WriteString(" ")
					break
				}

			}
			sentance.WriteString("\r\n\r\n")

			writer.WriteString(sentance.String())
		}

		writer.Flush()
		file.Close()
	}
	return usedWords
}

func TestDoSearch(t *testing.T) {
	//Setup test folder
	tmpDir, err := ioutil.TempDir("", "testing")
	if err != nil {
		t.Logf("Unable to create Temp Folder: %s", err.Error())
		t.FailNow()
	}

	//Create test files
	usedWords := generateTestData(tmpDir)

	//Setup test Server
	cfg := Config{data: tmpDir}
	s := &Server{config: cfg}

	//Do test
	testWord := usedWords[r.Intn(len(usedWords))] + ""
	results := s.DoSearch(testWord)
	resultLength := len(results.Hits)

	//Verify results
	assert.Equal(t, 1, resultLength)

	//Cleanup
	if err := os.RemoveAll(tmpDir); err != nil {
		t.Logf("Error removing temp directory['%s']: %s", tmpDir, err.Error())
	}
}

func TestcheckFile(t *testing.T) {

}
