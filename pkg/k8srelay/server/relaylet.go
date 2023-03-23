/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
)

const myclientcert = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURRakNDQWlxZ0F3SUJBZ0lJWDNENmkyZDFJNjh3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TXpBek1EY3lNekUzTlRKYUZ3MHlOREF6TURrd01UTTRNekphTURZeApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sc3dHUVlEVlFRREV4SmtiMk5yWlhJdFptOXlMV1JsCmMydDBiM0F3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRREVDejdWRldkSWg0bEIKSngyd1FId0hRMWpHRE9SdzZZZzdyRFk1S2xYZWhMcklyclMyY2IxcDdnRGZyMkc0dEJ4blpJOC9GR3JwT1VmQQpOVFRrMTdFRk13dnJWNHFIZjBxNjJXUGFPRHYxT3Zhdk1wRTFLUk1rVURBbEJiUWYrQzhNS1NxaXEyRWdjK052Cit3YVBwMU40STg5dTJZa2VoSTBYQTk2cWd6QmJJVWlJSlpqcFUyVUhOMWFtSVU3d0xXL1dybW1UN00xdFRtcncKL25nSmZiUW1IRFhwTHA4SjQwT3pWbEJZSm1YTWx5VHZ0VGowQXRRT1hSMiswMEtETnJsaXZRV01GbmRZZ3lQTQpZbVdheERkSUk1c0gvU0N6UjJDekxXbnh2WTltSXk0WVVvWEhvZTE5RDN5R0QrRHZsbDFOWHBGbkxISWs3THRXCnRsM2Z0b3U1QWdNQkFBR2pkVEJ6TUE0R0ExVWREd0VCL3dRRUF3SUZvREFUQmdOVkhTVUVEREFLQmdnckJnRUYKQlFjREFqQU1CZ05WSFJNQkFmOEVBakFBTUI4R0ExVWRJd1FZTUJhQUZBdVEyT1psYi94ZkZYMldadnd3cnhHKwpuK0laTUIwR0ExVWRFUVFXTUJTQ0VtUnZZMnRsY2kxbWIzSXRaR1Z6YTNSdmNEQU5CZ2txaGtpRzl3MEJBUXNGCkFBT0NBUUVBbHNoV3hnS1NMczA0R3NyUWM1anVBOTRGc3NVREEyM28wTVVia04zbTN0bXhpN1YrYy9zMzFTekoKbzBDQXJnRUtZU09GS0VPdnQ5RGRlTHBnRE1GMlVaMUdlVkdnUlV0TlU2QjlEM01Yb1RoeG9DNHNERFQ4U1VWNQo1cWNGTHJpeUVtNEp1dWdSZ1plN0Nzb2VJWUVubXNCYnEwK3dCWERIemZhTUE5WkplTmt1eHBtMXJiZmJDZkpHClQ0Q0tQMWRJc0drL0hjSzNUSUFZTnorRXNMRWdGSzJUbjZkR3dVdzFLMS9zMU9uVTJxbTdGVm1LRzVuSlVwbGoKd3FsNlF5Q1B0aTlKOUNCdnVpYm1LeXd0ajZ0dUZ3UnhLTjhDRWVJM2E1M2dGdGFvTFVLd2JRcjdCb0dRQ3B1awo2dmdkdHZ6VkdIMmdMRUhPYTlWazNsVFdhemxSK2c9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="

const myclientkey = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBeEFzKzFSVm5TSWVKUVNjZHNFQjhCME5ZeGd6a2NPbUlPNncyT1NwVjNvUzZ5SzYwCnRuRzlhZTRBMzY5aHVMUWNaMlNQUHhScTZUbEh3RFUwNU5leEJUTUw2MWVLaDM5S3V0bGoyamc3OVRyMnJ6S1IKTlNrVEpGQXdKUVcwSC9ndkRDa3FvcXRoSUhQamIvc0dqNmRUZUNQUGJ0bUpIb1NORndQZXFvTXdXeUZJaUNXWQo2Vk5sQnpkV3BpRk84QzF2MXE1cGsrek5iVTVxOFA1NENYMjBKaHcxNlM2ZkNlTkRzMVpRV0NabHpKY2s3N1U0CjlBTFVEbDBkdnROQ2d6YTVZcjBGakJaM1dJTWp6R0psbXNRM1NDT2JCLzBnczBkZ3N5MXA4YjJQWmlNdUdGS0YKeDZIdGZROThoZy9nNzVaZFRWNlJaeXh5Sk95N1ZyWmQzN2FMdVFJREFRQUJBb0lCQVFDSWlrQ0h6bkZ5RFp3bAorYVZ1MVdyTThEWUxNbjJFdXRJOHBYUGFtc2JWeFdJR1ZjL2RaaGlEaDlXcDlZKzlRZ3lxWWxwMmw3VGluUmVCCklrMmx1U2c4czlId1pyZEFLZ01WWWtWdWZrNGNQYVlFWGRiT3pMM2RROUJVYU1XTW9xRzUrWTROWUFtMHZhSkIKb3ZkdDVCOTVoSTk3NkJ6ZFdYcWU5ZjRHaW5xS2JNcTlPcFZNQkgzdmZWWWtMc3hMUndyekRUMnB5MGZxMzQvWApqM2RzSE45dFlHelM3d1A2Y2RaMjloSi9EWDBsYTNKejF4c3I2M0I0b2dzZk52ZTFSUTNKNXlOY2E5Wk11OUtrCnFITWVWamFYK21BWUYrSWJZUWlJZXpjenVzaEFJZ1dPTFp4Skczc3hIMmJVRXdIbjJLZldRbTEreDZlTHhOZTIKSzhJeTFUZFJBb0dCQU9pTmtzUGZvM3JTYm9uSWJJTVNXTVovZHVYTnhBdzFOM3VINTNrbFl2aU1aRU9CZXRHVwptM29YRXd4RFVGT3JKNTE2VFpadS9ETVhpeG1UUHROaWo3anFNU3FzWi94THQ3Q2d2anhGci93dnRZakVkOHRGClc3OUE1em05YVd4VHo1TFg0UW9rTUpkUGFCMGgvQzdRZ2tESG44cGtmRW03SUxOd2MxM3REQkFsQW9HQkFOZlAKVmZpdExvYnJSRUd6REY3VFhRVUtCVGdrcXhrUjNXN09rWHhFRENGLytmSlVoYTNCRElYay9hT3pOQm9STUt0MgpyTUVCM2lDcDV3cy9LSm1TUTl3c1EvZ25FTUVlajlBRnFQUE9WcXRtSXNzc1lzb3ladmtwRXN5a2dvMlZSYmRJCktrNzVVN1g2dHMxSC9aQ1Q2V1Nyc052RU1vYVBqLzA1OWRqaG90OEZBb0dBQXVOdlJUdUwza1NxMXM5RWVjUksKa2Z5WFQzZGt6Zm9EUEdlTnVuVjZhemZqTHV0MnlRK2owcnBpcEM1WjJ4QXZKOGVUR3lFNXhMQ3dLNXNtbHAyTQp4M0V4TnlSNURpc3FsdWtJTTl0eHVpSWxrUk5Qb1ppMDhRVXZXZ28rT3ZnM2hjMWtvQ21lNk9JMW10Y0hPTldpCktJZlNOa05WUDkweEpNbHF4V25pVW5rQ2dZQk05MC8wOFdhdmpZWjVXKzdrZnNNbEFlN2NtQTlCVUtMRld2eDkKOGhMVmU3dmJsaE5hNVllZTFRMDBiYnYrTS9WRW9YMTVGRDV4TGNjTnRzZTNCWGdZTk4xRXlrSHFiZ2ppS3JLWAp5UllWNk1ZdDZiV040UzNpWEtpc3ZWc21QWDl3bjFjZmRVSkttNURJWTQxbUc2cFlVZmN3V2FlZlgrSDljTWRpClF3NkFOUUtCZ1FDSjFYaU1hZEhjUXdjT1l5a0o4TlE3aEdPTE5JMzVKWDF0YTFOVW5UeS9QZHdTZHhQTFBVU3YKeVdCaE9tV3NBYkhRL0QzcmlrVDJvcWZKTGQ3Mm4vY2Vlb1g3Ym1rSFQ1SDhaK3l4ZE9Ga0FibXlra3lueURPRApiWGFYYWxOTWpWVmVZNCtuVTNUaXA2dDRtekdTUC9Yck5GWWFFbyttaHE3ejNLelRxWEd6K1E9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="

const caCert = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJek1ETXdOekl6TVRjMU1sb1hEVE16TURNd05ESXpNVGMxTWxvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTEtpCkgrUXFSMDRuMHNzMUlrWDhVRTdFdXBub0xlWkt6eUhIcGxRcG5pOFpBMERYL1JjNVlpbzBmdVA2Ui9zN1c5TFoKdzZWZnFDWFcvUjdqT3Y4eDUxRlViMmRHVGQ1M0k5eWlFUTZGR2NhbXM0L0VEc1Q5UjhRcGhUVE5MUkI4NDBURwpaWlNNQ3Z2VHRkTzE0eXFpQUdnT09SOFgzVWRiYVg5V1UyZ2wrbm8vWSt6NnhCTm1vTnFQTStoQ3dIS3BHR21DCnBlTDdZNTd3dDVvVVZaWU9aaXB6NkxOcklWU0wvMWt6MHdFcDRNSk5VcythSzc3OWJ0N2d4N2xwYjBlUDN1YisKSTY4elFqaXMxRTBIalhjWTBCOUdZOVdQU000QW5qeWNrOExmcTB0VFg1NkdZYUtia0NOcWkrcThGclk5TjJUdgo0NG52NFQ3N2hVb2xpOENNdDNjQ0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZBdVEyT1psYi94ZkZYMldadnd3cnhHK24rSVpNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBRi96NlpEakRrS0g0MGRrM0lVMQpXRzVSSUFoMlZJeG1mR1V6Nm54dU1BSnRrMVZkMDZaSXovYzh6RzBVU0VpREtiN1hNR003RTZtM3NvMFlMM21VClZlVnluVnZRSnhTL3JPSG5kMXB5N2UxKzU3a1Z3aEpCbjR1UVlMRkFwdTFnNlloL1VoSEc0bWgrQ1VJQ2dFU0wKWDlYZUFKVmQ1L1pFazFRbEpUWWd2UGd3U2djLzlVaHNIQ2ZHUERSZ0dwenhvb24rbHhzc2h6S3hCa2hUMTJjVApYUlcvKytpdENuRTAyYXh3ekY1VXBhcXpiNk53UzBmZXJUMi9nVTI1cllaeWFrWStZbTcwbStxbm5adFYzLzduCnRmT0xFWDJWSUtzWlI1OGxIWWN4UGxtMUlNc0I1djRSeXoxR0Nod1dvRng1c3loQjFtNE9Fd2JKaERtSnBma2QKRFowPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="

var relaylet Relaylet

type Relaylet struct {
	client *http.Client
}

func (t *Relaylet) init() {
	caCert, err := base64.StdEncoding.DecodeString(caCert)
	if err != nil {
		log.WithError(err).Fatalf("Bad b4 %s", err.Error())
	}

	clientCert, err := base64.StdEncoding.DecodeString(myclientcert)
	if err != nil {
		log.WithError(err).Fatalf("Bad b4 %s", err.Error())
	}
	clientKey, err := base64.StdEncoding.DecodeString(myclientkey)
	if err != nil {
		log.WithError(err).Fatalf("Bad b4 %s", err.Error())
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		panic("not ok")
	}
	cert, cerr := tls.X509KeyPair(clientCert, clientKey)
	if cerr != nil {
		panic("bad keypait")
	}

	t.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

}

const debug = false

func (t *Relaylet) DoCall(c *gin.Context) {
	if debug {
		c.JSON(202, "HELLO")
		return
	}
	hostBase := "https://kubernetes.docker.internal:6443"
	parse, err := url.Parse(c.Request.RequestURI)
	log.Infof(parse.User.String())
	log.Infof("URI %s", c.Request.URL.String())
	if err != nil {
		panic(err)
	}
	fullURL := fmt.Sprintf("%s%s", hostBase, parse.Path)

	relayreq, err := http.NewRequest(c.Request.Method, fullURL, c.Request.Body)
	if err != nil {
		panic(err)
	}

	resp, err := t.client.Do(relayreq)
	var respBody []byte

	log.Infof("Got resp status %d - len %d", resp.StatusCode, resp.ContentLength)
	for k, v := range resp.Header {
		log.Infof("%s = %s ", k, v[0])
		c.Header(k, v[0])
	}
	if resp.StatusCode < 300 {
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if len(respBody) > 0 {
			n, err := c.Writer.Write(respBody)
			if err != nil {
				log.WithError(err).Errorf("error writing body len %d err=%s", n, err.Error())
			}
		}
		c.Writer.Flush()
	}

	c.Status(resp.StatusCode)

}
