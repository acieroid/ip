package main

import (
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	//"strings"
)

const src = `
<!DOCTYPE html>
<html>
    <head>
        <meta http-equiv="Content-Type" content="text/html">
        <style type="text/css">
            body {
                text-align: center;
            }
        </style>
        <title>What's my IP?</title>
    </head>
    <body>
        <h1>What's my IP?</h1>
        {{ range .IPs }}
            <p>{{ . }}</p>
        {{ end }}
        <p>
(<a href="?raw=1">text-only</a>/<a href="http://paste.awesom.eu/Ffug&hl=go">code</a>)
                </p>
    </body>
</html>
`

var (
	port  = flag.String("port", "8085", "Listening HTTP port")
	host  = flag.String("host", "localhost", "Listening HTTP host")
	index = template.Must(template.New("index").Parse(src))
)

func getIPs(r *http.Request) []string {
	ips := make([]string, 1)

	// If no proxy:
	//  return strings.Split(r.RemoteAddr, ":")[0]
	// Proxy following RFC 7239: (not tested)
	//  Forwarded: for=xx.xx.xx.xx;
	// return strings.Split(r.Header.Get("Forwarded"), "=")[1]
	if r.URL.Path == "/" {
		ips = append(ips, r.Header.Get("X-Forwarded-For"))
	} else {
		// requesting DNS A/AAAA query
		if res, err := net.LookupIP(r.URL.Path[1:]); err != nil {
			ips = append(ips, err.Error())
		} else {
			for _, ip := range res {
				ips = append(ips, ip.String())
			}
		}
	}

	return ips
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("raw") != "" {
		for _, ip := range getIPs(r) {
			w.Write([]byte(ip + "\n"))
		}
	} else {
		s := struct{ IPs []string }{getIPs(r)}
		if err := index.Execute(w, &s); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/", handler)
	log.Println("Launching on http://:" + *host + *port)
	log.Fatal(http.ListenAndServe(*host+":"+*port, nil))
}
