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
		outFile                             *os.File
		nReader                             *bufio.Reader
		nWriter                             *bufio.Writer
		regFirstLineString                  string         = `^\s*(\d+)\s+(\d+)\s*$`
		regFloatsString                     string         = `^\s*(-*\d+\.\d+\w*[-\+]{0,1}\d*)\s+(-*\d+\.\d+\w*[-\+]{0,1}\d*)\s+(-*\d+\.\d+\w*[-\+]{0,1}\d*)`
		regIntsString                       string         = `^\s*(\d+)\s+(\d+)\s+(\d+)`
		regSingleIntString                  string         = `^\s*\d\s*$`
		regFirstLine                        *regexp.Regexp = regexp.MustCompile(regFirstLineString)
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
	go writeFloats(nWriter, floatChan, done)
	go writeInts(nWriter, intChan, done)

	// Open infile for reading
	if inFile, err = os.Open(inFileName); err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()
	nReader = bufio.NewReader(inFile)

	outFile, err = os.Create(baseFileName + ".plt")
	defer outFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	nWriter = bufio.NewWriter(outFile)
	defer nWriter.Flush()

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
		if regRes = regFloats.FindStringSubmatch(line); len(regRes) != 0 {
			// Found first line, write to file
			if _, err = nWriter.WriteString(regRes[1] + "\t" + regRes[2] + "\n"); err != nil {
				log.Fatalf("Can't write to %v with error %v\n", "output", err)
			}
			nWriter.Flush()
		} else if regRes = regFirstLine.FindStringSubmatch(line); len(regRes) != 0 {
			// Found floats, send them to coords file writing
			floatChan <- regRes
		} else if regRes = regInts.FindStringSubmatch(line); len(regRes) != 0 {
			// Found ints, send them to idxs file writing
			intChan <- regRes
		} else if regRes = regSingleInt.FindStringSubmatch(line); len(regRes) != 0 {
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

func writeFloats(nWriter *bufio.Writer, floatChan chan []string, done chan struct{}) {
	
	var (
		nums []string
		err  error
	)
	defer log.Println(len(nums))
	// Write to file
	for nums = range floatChan {
		if len(nums) == 0 {
			log.Fatal(nums)
		}
		if _, err = nWriter.WriteString(nums[1] + "\t" + nums[2] + "\t" + nums[3] + "\n"); err != nil {
			log.Fatalf("Can't write to %v with error %v\n", "output", err)
		}
		nWriter.Flush()
	}

	// Send end signal
	done <- struct{}{}
}

func writeInts(nWriter *bufio.Writer, intChan chan []string, done chan struct{}) {
	var (
		nums       []string
		err        error
		n1, n2, n3 int64
	)

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
			log.Fatalf("Can't write to %v with error %v\n", "output", err)
		}
		nWriter.Flush()
	}

	// Send end signal
	done <- struct{}{}
}
