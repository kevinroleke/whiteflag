package main

import (
	"bufio"
	"encoding/json"
	"os"
	"fmt"
	"github.com/lixiangzhong/dnsutil"
	"github.com/kevinroleke/go-domain-util/domainutil"
	"regexp"
	"strings"
	"errors"
	"sync"
)

var ( 
	TypeCNAME uint16 = 5
	TypeNS uint16 = 2
)

// https://stackoverflow.com/a/34388102
func JsonLXL(fname string, cb func(map[string]interface{}, dnsutil.Dig) ([]string, error), dig dnsutil.Dig, ch chan []string, wg *sync.WaitGroup) error {
	f, err := os.Open(fname)
	HandleErr(err, "*")

	s := bufio.NewScanner(f)
	for s.Scan() {
		var v map[string]interface{}
		err := json.Unmarshal(s.Bytes(), &v)
		if err != nil {
			wg.Done()
			return err
		}

		hit, err := cb(v, dig)
		if err != nil {
			// fmt.Println(err)
		}

		if hit != nil {
			ch <- hit
		}
	}

	wg.Done()

	if s.Err() != nil {
		return err
	}

	return nil
}

func CNAMELine(v map[string]interface{}, dig dnsutil.Dig) ([]string, error) {
	if v["type"] != "cname" {
		// return nil, errors.New("Bad line")
		return nil, nil
	}

	cto := strings.TrimRight(v["value"].(string), ".")
	domain := v["name"].(string)

	dig.SetDNS("8.8.8.8")
	return cnameCheck(cto, domain, dig)
}

func cnameCheck(cto string, domain string, dig dnsutil.Dig) ([]string, error) {
	msg, err := dig.GetMsg(TypeNS, cto)
	if err != nil {
		return nil, err
	}

	status, err := extractStatus(fmt.Sprint(msg))
	if err != nil {
		return nil, err
	}

	if status == "SERVFAIL" {
		// for _, item := range(msg.Answer) {
		// 	f := strings.Fields(fmt.Sprint(item))
		// 	nxCheck(f[len(f)-1], cto, dig)
		// }

		return hit("CNAME DOMAIN SERVFAIL", domain, cto), nil
	} else if status == "REFUSED" {
		return hit("CNAME DOMAIN REFUSED NO RECORDS", domain, cto), nil
	} else if status == "NXDOMAIN" {
		if domainutil.HasSubdomain(cto) {
			msg, err := dig.GetMsg(TypeNS, domainutil.Domain(cto))
			if err != nil {
				return nil, err
			}

			status, err := extractStatus(fmt.Sprint(msg))
			if err != nil {
				return nil, err
			}

			if status == "NXDOMAIN" {
				return hit("CNAME NXDOMAIN", domain, domainutil.Domain(cto)), nil
			}
		} else {
			return hit("CNAME NXDOMAIN", domain, cto), nil
		}
	}

	return nil, nil
}

func extractStatus(msg string) (string, error) {
	pattern := `status: (?P<a>.*), id`
	r := regexp.MustCompile(pattern)

	matches := r.FindStringSubmatch(msg)

	if len(matches) < 2 {
		return "", errors.New("Couldn't find status in DIG msg")
	}

	return matches[1], nil
}

func NSLine(v map[string]interface{}, dig dnsutil.Dig) ([]string, error) {
	if v["type"] != "ns" {
		// return nil, errors.New("Bad line")
		return nil, nil
	}

	ns := strings.TrimRight(v["value"].(string), ".")
	domain := v["name"].(string)

	// Check if the NS are NXDOMAIN
	dig.SetDNS("8.8.8.8")
	return nxCheck(ns, domain, dig)

	// Check if the NS refuse to resolve the domain
	dig.SetDNS(ns)
	return servfailCheck(domain, ns, dig)
}

func nxCheck(domain string, orig string, dig dnsutil.Dig) ([]string, error) {
	msg, err := dig.GetMsg(TypeNS, domain)
	if err != nil {
		return nil, err
	}

	status, err := extractStatus(fmt.Sprint(msg))
	if err != nil {
		return nil, err
	}

	if !domainutil.HasDomain(domain) {
		return nil, errors.New("bad domain")
	}

	if status == "NXDOMAIN" {
		if domainutil.HasSubdomain(domain) {
			return nxCheck(domainutil.Domain(domain), orig, dig)
		} else {
			return hit("NXDOMAIN NAMESERVER", orig, domain), nil
		}
	}

	return nil, nil
}

func servfailCheck(domain string, ns string, dig dnsutil.Dig) ([]string, error) {
	msg, err := dig.GetMsg(TypeNS, domain)
	if err != nil {
		return nil, err
	}

	status, err := extractStatus(fmt.Sprint(msg))
	if err != nil {
		return nil, err
	}

	if !domainutil.HasDomain(domain) {
		return nil, errors.New("bad domain")
	}

	if status == "SERVFAIL" {
		return hit("SERVFAIL", domain, ns), nil
	} else if status == "REFUSED" {
		return hit("REFUSED NO RECORDS", domain, ns), nil
	}

	return nil, nil
}

func hit(typ string, domain string, ns string) []string {
	return []string{typ, domain, ns}
}