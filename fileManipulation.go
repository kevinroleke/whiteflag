package main

import (
	"io"
	"bufio"
	// "fmt"
	"os"
	"bytes"
	"strconv"
	"path/filepath"
	"fmt"
	"errors"
	"strings"
	"encoding/json"
)

func DeleteWildcard(glob string) error {
	files, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}

	return nil
}

func SplitFile(lineCount int, workerCount int, fname string, prefix string) error {
	// fmt.Println(fname)
	f, err := os.Open(fname)
	HandleErr(err, "*")

	sc := bufio.NewScanner(f)
    i := 0
	chunkSize := lineCount / workerCount
	remainder := lineCount % workerCount
	chunkInd := 0

	var (
		// flush []string
		chunkWriters []*bufio.Writer
		chunkFiles []*os.File
	)

	if remainder > 0 {
		workerCount++
	}

	for ind := 0; ind < workerCount; ind++ {
		file, err := os.OpenFile(prefix + strconv.Itoa(ind), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		datawriter := bufio.NewWriter(file)

		chunkFiles = append(chunkFiles, file)
		chunkWriters = append(chunkWriters, datawriter)
	}

	for sc.Scan() {
        i++
        chunkWriters[chunkInd].WriteString(sc.Text() + "\n")

		if i % 100 == 0 {
			chunkWriters[chunkInd].Flush()
		}

		if i >= chunkSize {
			chunkWriters[chunkInd].Flush()
			chunkFiles[chunkInd].Close()

			chunkInd++
			i = 0
			
			if chunkInd >= workerCount {
				return nil
			}
		}
    }

	if remainder > 0 {
		chunkWriters[chunkInd].Flush()
		chunkFiles[chunkInd].Close()
	}

	return nil
}

func LoadHits() {
	if _, err := os.Stat("hits.json"); errors.Is(err, os.ErrNotExist) {
		return
	}

	f, err := os.Open(HitsFile)
	HandleErr(err, "*")

	s := bufio.NewScanner(f)
	for s.Scan() {
		var v map[string]interface{}
		err := json.Unmarshal(s.Bytes(), &v)
		// HandleErr(err, "*")
		if err != nil {
			continue
		}
		
		Hits = append(Hits, v)
	}

	fmt.Printf("Loaded %d hits\n", len(Hits))
}

func SaveHits() {
	f, err := os.OpenFile("hits.json", os.O_CREATE|os.O_WRONLY, 0644)
	HandleErr(err, "*")

	datawriter := bufio.NewWriter(f)
 
	for _, data := range Hits {
		line := new(bytes.Buffer)
		for k, v := range data {
			fmt.Fprintf(line, "\"%s\": \"%s\", ", strings.Trim(k, `"`), strings.Trim(v.(string), `"`))
		}
		stringified := fmt.Sprintf("{%s}\n", strings.TrimSuffix(line.String(), ", "))
		datawriter.WriteString(stringified)
	}
 
	datawriter.Flush()
	f.Close()
}

// https://stackoverflow.com/a/52153000
func LineCounter(r io.Reader) (int, error) {
    var count int
    const lineBreak = '\n'

    buf := make([]byte, bufio.MaxScanTokenSize)

    for {
        bufferSize, err := r.Read(buf)
        if err != nil && err != io.EOF {
            return 0, err
        }

        var buffPosition int
        for {
            i := bytes.IndexByte(buf[buffPosition:], lineBreak)
            if i == -1 || bufferSize == buffPosition {
                break
            }
            buffPosition += i + 1
            count++
        }
        if err == io.EOF {
            break
        }
    }

    return count, nil
}

// petergrep
func GrepFile(file string, pat []byte) int64 {
    patCount := int64(0)
    f, err := os.Open(file)
    if err != nil {
        fmt.Println(err.Error())
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        if bytes.Contains(scanner.Bytes(), pat) {
            patCount++
        }
    }
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
    return patCount
}