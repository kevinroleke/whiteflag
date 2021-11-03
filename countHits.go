package main

import (
	"fmt"
	"sort"
	"github.com/kevinroleke/go-domain-util/domainutil"
)

// this is WAY too slow
func CountHits() error {
	// var chits = []map[string]interface{}{}
	var nsCount = map[int]int64{}
	var cnameCount = map[int]int64{}
	LoadHits()

	for i, z := range Hits {
		fmt.Println(1)
		domain := domainutil.Domain(z["name"].(string))
		nsTotal := GrepFile("ns.json", []byte(domain))
		fmt.Println(2)
		cnameTotal := GrepFile("cname.json", []byte(domain))

		nsCount[i] = nsTotal
		cnameCount[i] = cnameTotal
	}

	fmt.Println(nsCount)
	fmt.Println(cnameCount)

	sort.Slice(nsCount, func(i1, i2 int) bool {
		return nsCount[i1] > nsCount[i2]
	})

	sort.Slice(cnameCount, func(i1, i2 int) bool {
		return cnameCount[i1] > cnameCount[i2]
	})

	fmt.Println(nsCount)
	fmt.Println(cnameCount)

	return nil
}