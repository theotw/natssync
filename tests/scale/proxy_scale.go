/*
 * Copyright (c) The One True Way 2022. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Printf("MAX Procs: %d \n",runtime.GOMAXPROCS(0))

	proxyStr := "http://proxylet:@localhost:30080"

	proxyURL, _ := url.Parse(proxyStr)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.DefaultTransport.(*http.Transport).Proxy= http.ProxyURL(proxyURL)

	var wg sync.WaitGroup
	t1:=time.Now()
	threads:=50
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
	t2:=time.Now()
	diff:=t2.Unix() - t1.Unix()
	fmt.Printf("** Total Time %d \n",diff)
}
func DoBunchOGets(tag,n int){
	for i:=0;i<n;i++ {
		fmt.Printf("%d - %d \n",tag, i)
		DoGet()
	}
}
func DoGet() {

	resp, err := http.Get("https://192.168.65.4/john-work.jpeg")

	if err != nil {
		fmt.Printf("Error %s \n", err.Error())
	}else {
		if resp.StatusCode != 200 {
			fmt.Printf("Code %s \n", resp.Status)
		}else{
			_, err := ioutil.ReadAll(resp.Body)
			if err != nil{
				fmt.Printf("Error reading body %s",err.Error())
			}

		}
		//fmt.Printf("Success %s \n", resp.Status)
	}

}
