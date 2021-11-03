package main

import (
	"regexp"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"errors"
	"fmt"
	"compress/gzip"
)

func DownloadGunzip(url string, dst string) {
	out, err := os.Create(dst)
	HandleErr(err, "*")
	defer out.Close()

	resp, err := http.Get(url)
	HandleErr(err, "*")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		HandleErr(errors.New("URL returned non-OK status."), "*")
	}

	reader, err := gzip.NewReader(resp.Body)
	defer reader.Close()

	_, err = io.Copy(out, reader)
	HandleErr(err, "*")
}

func getArchiveName(record string) string {
	url := "https://opendata.rapid7.com/sonar.fdns_v2/"

	resp, err := http.Get(url)
	HandleErr(err, "*")

	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	HandleErr(err, "*")

	pattern := `nofollow\">(?P<fn>.*fdns_` + record + `\.json\.gz)`

	r := regexp.MustCompile(pattern)

	matches := r.FindStringSubmatch(string(html))

	return url + matches[1]
} 

func DownloadCNAME() {
	url := getArchiveName("cname")
	DownloadGunzip(url, "cname.json")
}

func DownloadNS() {
	url := getArchiveName("ns")
	DownloadGunzip(url, "ns.json")
}

func CheckFDNS() {
	if _, err := os.Stat("cname.json"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Downloading FDNS CNAME data from Rapid7 OpenData. This may take a while.")
		DownloadCNAME()
		fmt.Println(">> cname.json\nDone!")
	}

	if _, err := os.Stat("ns.json"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Downloading FDNS NS data from Rapid7 OpenData. This may take a while.")
		DownloadNS()
		fmt.Println(">> ns.json\nDone!")
	}
}