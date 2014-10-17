package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	client := &http.Client{}
	data := "{\"consumer_key\": \"22838-50555f6efec6293dddbdc4ae\", \"redirect_uri\": \"gulyasm-personal-stat:authorizationFinished\"}"
	buf := bytes.NewBufferString(data)
	req, err := http.NewRequest("POST", "https://getpocket.com/v3/oauth/request", buf)
	if err != nil {
		fmt.Println("Error")
	}

	req.Header.Add("X-Accept", "application/json")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error doing the request.")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	type CodeMessage struct {
		Code, Status string
	}
	code := &CodeMessage{}
	json.Unmarshal(body, code)
	fmt.Println(code.Code)

	type GetData struct {
		AccessToken string `json:"access_token"`
		State       string `json:"state"`
		ConsumerKey string `json:"consumer_key"`
		Sort        string `json:"sort"`
		DetailType  string `json:"detailType"`
	}

	getdata := &GetData{
		AccessToken: code.Code,
		ConsumerKey: "22838-50555f6efec6293dddbdc4ae",
		State:       "unread",
		Sort:        "newest",
		DetailType:  "simple",
	}
	b, err := json.Marshal(getdata)
	if err != nil {
		fmt.Println("Error serialazing getdata")
		return
	}
	req, err = http.NewRequest("POST", "http://getpocket.com/v3/get", bytes.NewReader(b))
	if err != nil {
		fmt.Println("Error")
	}

	req.Header.Add("X-Accept", "application/json")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error doing the request.")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

}
