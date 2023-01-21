package bybit_client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

var url = "https://api-testnet.bybit.com"

type Client struct {
	apiKey     string
	apiSecret  string
	recvWindow int
	client     *http.Client
	isDebug    bool
}

func NewClient(apiKey string, apiSecret string, recvWindow int, client *http.Client, isDebug bool) *Client {
	return &Client{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		recvWindow: recvWindow,
		client:     client,
		isDebug:    isDebug,
	}
}

func (c Client) GetRequest(params string, endPoint string) []byte {
	now := time.Now()
	unixNano := now.UnixNano()
	timeStamp := unixNano / 1000000
	signature := c.getSignature(timeStamp, params)
	request := c.makeRequest("GET", url+endPoint+"?"+params, nil)
	c.setHeader(request, signature, timeStamp)
	if c.isDebug {
		c.dumpRequest(request)
	}
	response := c.doRequest(request)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if c.isDebug {
		c.debugResponse(now, endPoint, response, body)
	}
	return body
}

func (c Client) PostRequest(params interface{}, endPoint string) []byte {
	now := time.Now()
	unixNano := now.UnixNano()
	timeStamp := unixNano / 1000000
	jsonData, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}
	signature := c.getSignature(timeStamp, string(jsonData[:]))
	request := c.makeRequest("GET", url+endPoint, bytes.NewBuffer(jsonData))
	c.setHeader(request, signature, timeStamp)
	if c.isDebug {
		c.dumpRequest(request)
	}
	response := c.doRequest(request)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if c.isDebug {
		c.debugResponse(now, endPoint, response, body)
	}
	return body
}

func (c Client) makeRequest(method string, url string, body io.Reader) *http.Request {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatal(err)
	}
	return request
}

func (c Client) doRequest(request *http.Request) *http.Response {
	response, err := c.client.Do(request)
	if err != nil {
		panic(err)
	}
	return response
}

func (c Client) debugResponse(now time.Time, endPoint string, response *http.Response, body []byte) {
	elapsed := time.Since(now).Seconds()
	fmt.Printf("\n%s took %v seconds \n", endPoint, elapsed)
	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	fmt.Println("response Body:", string(body))
}

func (c Client) dumpRequest(request *http.Request) {
	reqDump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Request Dump:\n%s", string(reqDump))
}

func (c Client) setHeader(request *http.Request, signature string, timeStamp int64) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-BAPI-API-KEY", c.apiKey)
	request.Header.Set("X-BAPI-SIGN", signature)
	request.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timeStamp, 10))
	request.Header.Set("X-BAPI-SIGN-TYPE", "2")
	request.Header.Set("X-BAPI-RECV-WINDOW", strconv.Itoa(c.recvWindow))
}

func (c Client) getSignature(timeStamp int64, params string) string {
	hmac256 := hmac.New(sha256.New, []byte(c.apiKey))
	hmac256.Write([]byte(strconv.FormatInt(timeStamp, 10) + c.apiKey + strconv.Itoa(c.recvWindow) + params))
	signature := hex.EncodeToString(hmac256.Sum(nil))
	return signature
}
