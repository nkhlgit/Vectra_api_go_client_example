//
// Author: Nikhil Singh
// Email: nikhil.eltrx@gmail.com
// Purpose: Pull hosts data from Vectra brain using API and sve the output json file defined in conf file..
// Usage: 
//   - configure get_hosts_conf.json
//   - 
//Compatiblity_tested: Python3, VEctra Brain: 7.1, API version : 2.2 :
//

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var err error

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type Conf struct {
	Vec_he          string `json:"vec_he"`
	Vec_api_token   string `json:"vec_api_token"`
	Max_page_number int    `json:"max_page_number"`
	Max_page_size   int    `json:"max_page_size"`
}

var conf Conf

func get_conf() {
	var homedir string
	homedir, err = os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	var conf_byte []byte
	fmt.Print(homedir)

	conf_file := homedir + "/get_hosts_conf.json"
	conf_byte, err = ioutil.ReadFile(conf_file)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	err = json.Unmarshal(conf_byte, &conf)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
}

func main() {
	get_conf()
	max_page_number := conf.Max_page_number
	c := NewClient()
	ctx := context.Background()
	var result map[string]interface{}
	page := 1
	var req *http.Request
	req, err = c.Gethosts_prepare(ctx)
	if err != nil {
		log.Fatal("error from Gethosts: ", err)
	}
	for i := 0; i < max_page_number; i++ {
		err = c.Gethosts_send(req, page, &result)
		if err != nil {
			log.Fatal("error from Gethosts: ", err)
		}
		if i == 0 {
			result_count := result["count"]
			log.Printf("result_count= %v \n", result_count)
		}
		result_next := result["next"]
		if result_next == nil {
			log.Printf("Breaking out as next is:%v ", result_next)
			break
		}
		log.Printf("result_next= %v \n", result_next)
		page++
	}
}

type Client struct {
	BaseURL    string
	apiKey     string
	HTTPClient *http.Client
}

func NewClient() *Client {
	tr := &http.Transport{
		// This is the insecure setting, it should be set to false.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	vec_api_url := fmt.Sprintf("https://%s/api/v2.2", conf.Vec_he)
	host_url := vec_api_url + "/hosts"
	apikey := "Token " + conf.Vec_api_token
	return &Client{
		BaseURL: host_url,
		apiKey:  apikey,
		HTTPClient: &http.Client{
			Timeout:   time.Minute,
			Transport: tr,
		},
	}
}

func (c *Client) Gethosts_prepare(ctx context.Context) (*http.Request, error) {

	req, err := http.NewRequest("GET", c.BaseURL, nil)
	if err != nil {
		return req, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", c.apiKey)
	params := req.URL.Query()
	params.Add("active_traffic", "True")
	max_page_size := conf.Max_page_size
	params.Add("page_size", strconv.Itoa(max_page_size))
	params.Add("page", "1")
	req.URL.RawQuery = params.Encode()
	return req, err
}
func (c *Client) Gethosts_send(req *http.Request, page int, result *map[string]interface{}) error {
	params := req.URL.Query()
	params.Set("page", strconv.Itoa(page))
	req.URL.RawQuery = params.Encode()
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return errors.New(errRes.Message)
		}
		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}
	//if err = json.NewDecoder(res.Body).Decode(hosts); err != nil {
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	json.Unmarshal([]byte(bodyBytes), &result)
	return nil
}
