# WHITEFLAG

Scan the internet for misconfigured domain names.

## CNAME Takeover

A CNAME record is used to point a subdomain on website A, to another domain. 

But what happens when one is neglected? It leaves the possibility to be taken over. 
The domain it points to, may have expired, may be incorrect, may be pointing to a service like Github. 

## Nameserver Takeover

While most people use their hosting provider's nameservers, or ones of Cloudflare/a DNS service, some domain operators opt to use custom ones. Say your nameservers are set to an expired domain, test.in. This means when a user tries to query yourdomain.com, it contacts test.in asking for records. If an attacker registers test.in and sets up a DNS server on it, they decide what records exist on yourdomain.com, leading to a complete takeover. 

## Building

```shell
$ git clone git@github.com:kevinroleke/whiteflag.git
$ cd whiteflag
$ go build
```

## Usage

```shell
$ ./whiteflag --scan
$ ./whiteflag --cname
```

## TODO

- Add a fast counting and sorting system to hunt for big takeovers.
- Add a visualization
- Add a second filter that uses WHOIS to determine if an NXDOMAIN is really available for registration. 
