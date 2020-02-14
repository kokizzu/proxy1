package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

const LISTEN = `:20`
const CACHEDIR = `./cache/`

var cache *Cache 

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
}

func handleError(err error, w http.ResponseWriter) {
	fmt.Println(err.Error())
	w.WriteHeader(500)
	fmt.Fprintf(w, err.Error())
}

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)
	if req.URL.Scheme == `` {
		if req.URL.Port() == `443` {
			req.URL.Scheme = "https"
			req.URL.Host = req.URL.Hostname()
		} else {
			req.URL.Scheme = "http"
		}
	}
	
	fullUrl := req.URL.String()
	fmt.Println(fullUrl)
	
	client := &http.Client{}

	//http: Request.RequestURI can't be set in client requests.
	//http://golang.org/src/pkg/net/http/client.go
	req.RequestURI = ""

	delHopHeaders(req.Header)

	if req.Method == http.MethodGet && cache.has(fullUrl) {
		fmt.Println(fullUrl)
		content, err := cache.get(fullUrl)
		if err != nil {
			handleError(err, wr)
		} else {
			wr.Write(content)
		}
	} 
	
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}

	defer resp.Body.Close()
	log.Println(req.RemoteAddr, " ", resp.Status)
	
	if req.Method == http.MethodGet {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			handleError(err, wr)
			return
		}
		err = cache.put(fullUrl, body)
		if err != nil {
			fmt.Printf("Failed write into cache: %s\n", err)
		}
		wr.Write(body)
	}

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	if req.Method != http.MethodGet {
		io.Copy(wr, resp.Body)
	}
	
	
}

func main() {
	addr := flag.String("LISTEN", LISTEN, "The LISTEN of the application.")
	cacheDir := flag.String("CACHEDIR", CACHEDIR, "The CACHE DIRectory, please end with /")
	flag.Parse()
	var err error
	cache, err = CreateCache(*cacheDir)
	if err != nil {
		fmt.Println(`Unable to create cache directory: ` + *cacheDir)
		return
	}
	
	handler := &proxy{}
	
	log.Println("Starting proxy server on", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
