package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/brunetto/goutils/debug"
	"github.com/brunetto/goutils/readfile"
)

func main() {
	defer debug.TimeMe(time.Now())

	if len(os.Args) < 2 {
		log.Fatal("Please provide a .ver input file name")
	}

	var (
		err                                 error
		inFileName, baseFileName, ext, line string
		inFile                              *os.File
		nReader                             *bufio.Reader
		regFloatsString                     string         = `^\s*(-*\d+\.\d+\w*[-\+]{0,1}\d*)\s+(-*\d+\.\d+\w*[-\+]{0,1}\d*)\s+(-*\d+\.\d+\w*[-\+]{0,1}\d*)`
		regIntsString                       string         = `^\s*(\d+)\s+(\d+)\s+(\d+)`
		regSingleIntString                  string         = `^\s*\d`
		regFloats                           *regexp.Regexp = regexp.MustCompile(regFloatsString)
		regInts                             *regexp.Regexp = regexp.MustCompile(regIntsString)
		regSingleInt                        *regexp.Regexp = regexp.MustCompile(regSingleIntString)
		regRes                              []string
		done                                    = make(chan struct{})
		floatChan                               = make(chan []string)
		intChan                                 = make(chan []string)
		nLines                              int = 0
	)

	inFileName = os.Args[1]

	ext = filepath.Ext(inFileName)
	if ext != ".ver" {
		log.Fatal("You Must provide a .ver file")
	}

	baseFileName = strings.Trim(inFileName, ext)

	// Start goroutines
	go writeFloats("coords-"+baseFileName+".plt", floatChan, done)
	go writeInts("idxs-"+baseFileName+".plt", intChan, done)

	// Open infile for reading
	if inFile, err = os.Open(inFileName); err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()
	nReader = bufio.NewReader(inFile)

	// Scan lines
	for {
		if line, err = readfile.Readln(nReader); err != nil {
			if err.Error() != "EOF" {
				log.Fatal("Done reading with err", err)
			} else {
				fmt.Printf("Parsed %v lines\n", nLines)
				log.Println("Found end of file.")
			}
			break
		}

		// Feedback on parsing
		nLines += 1
		if nLines%100 == 0 {
			fmt.Printf("Parsed %v lines\r", nLines)
		}

		// Try to understand the line type
		if regRes = regFloats.FindStringSubmatch(line); regRes != nil {
			// Found floats, send them to coords file writing
			floatChan <- regRes
		} else if regRes = regInts.FindStringSubmatch(line); regRes != nil {
			// Found ints, send them to idxs file writing
			intChan <- regRes
		} else if regRes = regSingleInt.FindStringSubmatch(line); regRes != nil {
			// Found single int, do nothing
		} else {
			log.Fatal("Can't understand line ", line)
		}
	}

	// Close channels to send shutdown signal to goroutines
	close(floatChan)
	close(intChan)

	// Empty "done"channels to complete goroutines shutdown
	for idx := 0; idx < 2; idx++ {
		<-done
	}

}

func writeFloats(fileName string, floatChan chan []string, done chan struct{}) {
	var (
		outFile *os.File
		nWriter *bufio.Writer
		nums    []string
		err     error
	)
	
	// Open file for writing
	outFile, err = os.Create(fileName)
	defer outFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	nWriter = bufio.NewWriter(outFile)
	defer nWriter.Flush()

	// Write to file
	for nums = range floatChan {
		if _, err = nWriter.WriteString(nums[1] + "\t" + nums[2] + "\t" + nums[3] + "\n"); err != nil {
			log.Fatalf("Can't write to %v with error %v\n", fileName, err)
		}
	}
	// For some reason, flush does not work with defer here
	nWriter.Flush()
	// Send end signal
	done <- struct{}{}
}

func writeInts(fileName string, intChan chan []string, done chan struct{}) {
	var (
		outFile    *os.File
		nWriter    *bufio.Writer
		nums       []string
		err        error
		n1, n2, n3 int64
	)
	
	// Open file for writing
	outFile, err = os.Create(fileName)
	defer outFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	nWriter = bufio.NewWriter(outFile)
	defer nWriter.Flush()

	// Write to file decreasing ints by 1
	for nums = range intChan {
		// Parse Ints to be able to decrease them
		if n1, err = strconv.ParseInt(nums[1], 10, 64); err != nil {
			log.Fatal("Can't parse a float number in ", nums[1])
		}
		if n2, err = strconv.ParseInt(nums[2], 10, 64); err != nil {
			log.Fatal("Can't parse a float number in ", nums[2])
		}
		if n3, err = strconv.ParseInt(nums[3], 10, 64); err != nil {
			log.Fatal("Can't parse a float number in ", nums[3])
		}

		// Write decreased ints to file
		if _, err = nWriter.WriteString(strconv.FormatInt(n1-1, 10) + "\t" +
			strconv.FormatInt(n2-1, 10) + "\t" +
			strconv.FormatInt(n3-1, 10) + "\n"); err != nil {
			log.Fatalf("Can't write to %v with error %v\n", fileName, err)
		}

	}
	// For some reason, flush does not work with defer here
	nWriter.Flush()
	// Send end signal
	done <- struct{}{}
}
