package main

import (
	"fmt"
	"os"
	"strconv"
	"errors"
	"sync"
	"flag"
	"github.com/lixiangzhong/dnsutil"
)

var (
	ch chan []string
	wg *sync.WaitGroup
	Hits = []map[string]interface{}{}
	HitsFile string = "hits.json"
	inc int = 0
)

func HandleErr(err error, typ string) {
	if err != nil {
		switch typ {
		case "*":
			panic(err)
		}
	}
}

func mpCNAME(workerCount int) {
	prefix := "cname_"
	didSplit := true
	for i := 0; i < workerCount+1; i++ {
		if _, err := os.Stat(prefix + strconv.Itoa(i)); errors.Is(err, os.ErrNotExist) {
			didSplit = false
			DeleteWildcard(prefix + "*")
			break
		}
	}
	if _, err := os.Stat(prefix + strconv.Itoa(workerCount+2)); err == nil {
		DeleteWildcard(prefix + "*")
		didSplit = false
	}

	if !didSplit {
		f, err := os.Open("cname.json")
		HandleErr(err, "*")

		fmt.Println("Counting lines in file")
		lines, err := LineCounter(f)
		HandleErr(err, "*")
		fmt.Printf("%d lines of data\n", lines)

		fmt.Println("Splitting file into parts")
		err = SplitFile(lines, workerCount, "cname.json", prefix)
		HandleErr(err, "*")
		fmt.Println("done")
	}

	for i := 0; i < workerCount+1; i++ {
		var dig dnsutil.Dig
		wg.Add(1)
        go JsonLXL("cname_" + strconv.Itoa(i), CNAMELine, dig, ch, wg)
    }
}

func mpNS(workerCount int) {
	prefix := "ns_"
	didSplit := true
	for i := 0; i < workerCount+1; i++ {
		if _, err := os.Stat(prefix + strconv.Itoa(i)); errors.Is(err, os.ErrNotExist) {
			didSplit = false
			DeleteWildcard(prefix + "*")
			break
		}
	}
	if _, err := os.Stat(prefix + strconv.Itoa(workerCount+2)); err == nil {
		DeleteWildcard(prefix + "*")
		didSplit = false
	}

	if !didSplit {
		f, err := os.Open("ns.json")
		HandleErr(err, "*")

		fmt.Println("Counting lines in file")
		lines, err := LineCounter(f)
		HandleErr(err, "*")
		fmt.Printf("%d lines of data\n", lines)

		fmt.Println("Splitting file into parts")
		err = SplitFile(lines, workerCount, "ns.json", prefix)
		HandleErr(err, "*")
		fmt.Println("done")
	}

	for i := 0; i < workerCount+1; i++ {
		var dig dnsutil.Dig
		wg.Add(1)
        go JsonLXL("ns_" + strconv.Itoa(i), NSLine, dig, ch, wg)
    }
}

func result(v []string) {
	fmt.Printf("HIT (%s): %s @%s\n", v[0], v[1], v[2])

	r := make(map[string]interface{})
	r["type"] = v[0]
	r["name"] = v[2]
	r["orig"] = v[1]

	for _, v := range Hits {
		if v["name"] == r["name"] && v["orig"] == r["orig"] {
			return
		}
	}

	Hits = append(Hits, r)

	inc++
	if inc % 25 == 0 {
		SaveHits()
	}
}

func main() {
	scan := flag.Bool("scan", false, "Begin scanning for both cname and ns misconfigurations.")
	ns := flag.Bool("ns", false, "Begin scanning for only ns misconfigurations.")
	cname := flag.Bool("cname", false, "Begin scanning for only cname misconfigurations.")

    flag.Parse()

	if *scan || *ns || *cname {
		fmt.Println("Checking for FDNS data...")
		CheckFDNS()

		fmt.Println("Loading previous hits...")
		LoadHits()

		ch = make(chan []string)
		wg = &sync.WaitGroup{}

		if *scan || *cname {
			fmt.Println("Scanning CNAME on 32 threads...")
			mpCNAME(32)
		}

		if *scan || *ns {
			fmt.Println("Scanning NS on 32 threads...")
			mpNS(32)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		for v := range ch {
			result(v)
		}
	} else {
		fmt.Println("Please supply either --ns, --cname, or --scan.")
	}
}