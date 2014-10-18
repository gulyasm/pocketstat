package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

const (
	redirect_uri = "https://getpocket.com/connected_accounts"
)

func obtainCode() (string, error) {
	client := &http.Client{}
	data := "{\"consumer_key\": \"22838-50555f6efec6293dddbdc4ae\", \"redirect_uri\": \"gulyasm-personal-stat:authorizationFinished\"}"
	buf := bytes.NewBufferString(data)
	req, err := http.NewRequest("POST", "https://getpocket.com/v3/oauth/request", buf)
	if err != nil {
		return "", err
	}

	req.Header.Add("X-Accept", "application/json")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	type CodeMessage struct {
		Code, Status string
	}
	code := &CodeMessage{}
	json.Unmarshal(body, code)
	return code.Code, nil
}

type AccessTokenResponse struct {
	UserName    string `json:"username"`
	AccessToken string `json:"access_token"`
}

func obtainToken(code string) (string, error) {
	fmt.Println("Requesting token...")
	client := &http.Client{}
	url := "https://getpocket.com/v3/oauth/authorize"
	data := "{\"consumer_key\": \"22838-50555f6efec6293dddbdc4ae\", \"code\":\"%s\"}"
	data = fmt.Sprintf(data, code)
	fmt.Println(data)
	buf := bytes.NewBufferString(data)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Accept", "application/json")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp, err := client.Do(req)
	fmt.Println(resp.Status)
	if err != nil || resp.StatusCode != 200 {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	atr := &AccessTokenResponse{}
	json.Unmarshal(body, atr)
	return atr.AccessToken, nil
}

type Config struct {
	Code, Token string
}

func main() {
	config_path := ".pocketstat"
	configfile, err := ioutil.ReadFile(config_path)
	config := &Config{}
	if err != nil {
		fmt.Println(err)
	} else {
		json.Unmarshal(configfile, config)
	}
	var perm os.FileMode = 0777
	if config.Code == "" {
		code, err := obtainCode()
		if err != nil {
			fmt.Println("Major fuckup!")
		}
		config.Code = code
		access_token_url := fmt.Sprintf("https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s", code, redirect_uri)
		data, err := json.Marshal(config)
		fmt.Println(string(data))
		err = ioutil.WriteFile(config_path, data, perm)
		if err != nil {
			fmt.Println(err)
		}
		err = exec.Command("xdg-open", access_token_url).Start()
		return
	}
	access_token := config.Token
	if access_token == "" {
		access_token, err := obtainToken(config.Code)
		if err != nil {
			fmt.Println("Major fuckup at token.")
		}
		if access_token == "" {
			fmt.Println("Empty Access Token. Exiting...")
			return
		}
		config.Token = access_token
		data, err := json.Marshal(config)
		ioutil.WriteFile(config_path, data, perm)
	}

	type GetData struct {
		AccessToken string `json:"access_token"`
		State       string `json:"state"`
		ConsumerKey string `json:"consumer_key"`
		Sort        string `json:"sort"`
		DetailType  string `json:"detailType"`
	}

	getdata := &GetData{
		AccessToken: config.Token,
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
	req, err := http.NewRequest("POST", "http://getpocket.com/v3/get", bytes.NewReader(b))
	if err != nil {
		fmt.Println("Error")
	}

	client := http.Client{}
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
	fmt.Println(string(body))

}
