package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var domainsToAddresses map[string]string = map[string]string{}

type handler struct{}

func (this *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}

	msg.SetReply(r)
	msg.Authoritative = true

	domain := msg.Question[0].Name
	println(domain)

	address, ok := domainsToAddresses[domain]

	if ok {
		msg.Answer = append(msg.Answer, &dns.A{
			Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
			A:   net.ParseIP(address),
		})
	} else {

		if !query(domain, w, r) {
			safe(domain)
		}

		for key, element := range domainsToAddresses {
			fmt.Println("Key:", key, "=>", "Element:", element)
		}

	}

	w.WriteMsg(&msg)
}

func queryFull(dName string) string {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	ip, _ := r.LookupHost(context.Background(), dName)

	return ip[0]
}

func query(domainName string, w dns.ResponseWriter, r *dns.Msg) bool {
	println("Querying1 " + domainName)
	ips, err := net.LookupIP(domainName)
	if err != nil {
		if safe(domainName) {
			queryFull(domainName)
		}
		return false
	}

	println(ips)
	fmt.Printf("%s. IN A %s\n", domainName, ips[0].String())
	domainsToAddresses[domainName] = ips[0].String()

	msg := dns.Msg{}

	msg.SetReply(r)
	msg.Authoritative = true

	domain := msg.Question[0].Name

	address, ok := domainsToAddresses[domain]

	if ok {
		msg.Answer = append(msg.Answer, &dns.A{
			Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
			A:   net.ParseIP(address),
		})
	}

	w.WriteMsg(&msg)

	return true

}

func safe(url string) bool {
	var q = fmt.Sprint("http://127.0.0.1:5000?ip=", url)
	resp, err := http.Get(q)
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	input := "yes"
	if strings.Contains(sb, input) {
		queryFull(url)
		return true
	} else {
		println("Not safe")
		return false
	}

}

func main() {
	println("Started")
	srv := &dns.Server{Addr: ":" + strconv.Itoa(53), Net: "udp"}
	srv.Handler = &handler{}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}

}
