/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"sync"
)

func main() {
	fmt.Printf("MAX Procs: %d \n",runtime.GOMAXPROCS(0))
	proxyStr := "http://proxylet:@localhost:8080"
	proxyURL, _ := url.Parse(proxyStr)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.DefaultTransport.(*http.Transport).Proxy= http.ProxyURL(proxyURL)

	var wg sync.WaitGroup
	threads:=1000
	rounds:=10
	wg.Add(threads)
	for i:=0;i<threads;i++{
		x:=i
		go func() {
			DoBunchOGets(x,rounds)
			wg.Done()
		}()
	}

	wg.Wait()
}
func DoBunchOGets(tag,n int){
	for i:=0;i<n;i++ {
		fmt.Printf("%d - %d \n",tag, i)
		DoGet()
	}
}
func DoGet() {
	resp, err := http.Get("https://localhost/john-work.jpeg")
	if err != nil {
		fmt.Printf("Error %s \n", err.Error())
	}else {
		if resp.StatusCode != 200 {
			fmt.Printf("Code %s \n", resp.Status)
		}
		//fmt.Printf("Success %s \n", resp.Status)
	}

}
