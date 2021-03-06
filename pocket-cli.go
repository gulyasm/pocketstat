package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	redirect_uri = "https://getpocket.com/connected_accounts"
	perm         = 0777
)

type PocketStat struct {
	Id       bson.ObjectId            `bson:"_id"`
	Articles []map[string]interface{} `bson:"articles"`
	Count    int                      `bson:"count"`
	Time     int64                    `bson:"timestamp"'`
}

func openUrl(url string) (err error) {
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows", "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func obtainCode() (string, error) {
	client := &http.Client{}
	data, _ := json.Marshal(struct {
		ConsumerKey string `json:"consumer_key"`
		RedirectUri string `json:"redirect_uri"`
	}{
		ConsumerKey: "22838-50555f6efec6293dddbdc4ae",
		RedirectUri: "gulyasm-personal-stat:authorizationFinished",
	})
	buf := bytes.NewBuffer(data)
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
	client := &http.Client{}
	url := "https://getpocket.com/v3/oauth/authorize"
	data, _ := json.Marshal(struct {
		ConsumerKey string `json:"consumer_key"`
		Code        string `json:"code"`
	}{
		ConsumerKey: "22838-50555f6efec6293dddbdc4ae",
		Code:        code,
	})
	buf := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Accept", "application/json")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	atr := &AccessTokenResponse{}
	json.Unmarshal(body, atr)
	return atr.AccessToken, nil
}

type Config struct {
	Code, Token string
}

func insertIntoDb(articles []map[string]interface{}) error {
	sess, err := mgo.Dial("localhost")
	collection := sess.DB("pocket-stat").C("article-collections")
	doc := PocketStat{Id: bson.NewObjectId(), Articles: articles, Count: len(articles), Time: time.Now().Unix()}
	err = collection.Insert(doc)
	return err
}

func main() {
	var config_path = flag.String("config", "", "The path to the config file")
	var format = flag.String("format", "csv", "The format specifier. Has to be one of the following: csv | db")
	var v = flag.String("v", "elements", "The verbosity level. One of the following: elements | count")
	flag.Parse()

	if *config_path == "" {
		fmt.Println("config_path not given")
		return
	}

	configfile, err := ioutil.ReadFile(*config_path)
	config := &Config{}
	if err != nil {
		fmt.Println(err)
	} else {
		json.Unmarshal(configfile, config)
	}
	if config.Code == "" {
		code, err := obtainCode()
		if err != nil {
			fmt.Println("Major fuckup!")
		}
		config.Code = code
		access_token_url := fmt.Sprintf("https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s", code, redirect_uri)
		data, err := json.Marshal(config)
		err = ioutil.WriteFile(*config_path, data, perm)
		if err != nil {
			fmt.Println(err)
		}
		openUrl(access_token_url)
		fmt.Printf("Please authorize pocket-stat application in your browser. If the authorization page did not open, please open the following url in your browser: %s\n", access_token_url)
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
		ioutil.WriteFile(*config_path, data, perm)
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
	var f interface{}
	err = json.Unmarshal(body, &f)
	var message map[string]interface{} = f.(map[string]interface{})
	f = message["list"]
	var articles map[string]interface{} = f.(map[string]interface{})
	var articles_list []map[string]interface{}
	for _, v := range articles {
		var article map[string]interface{} = v.(map[string]interface{})
		articles_list = append(articles_list, article)
	}

	switch *format {
	case "db":
		insertIntoDb(articles_list)
	case "console":
		fmt.Println(len(articles_list))
	case "csv":
		switch *v {
		case "count":
			fmt.Printf("%d;%d\n", time.Now().Unix(), len(articles_list))
		case "elements":
			for _, a := range articles_list {
				fmt.Printf("%d;%s;%s;%s\n",
					time.Now().Unix(),
					a["time_added"],
					strconv.Quote(a["resolved_title"].(string)),
					a["word_count"])
			}
		}
	}
}
