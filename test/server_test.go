package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestServerDemo(t *testing.T) {
	http.HandleFunc("/get-timeout", getTimeout)
	http.HandleFunc("/getWithQuery", getWithQuery)
	http.HandleFunc("/post-with-cookies", postWithCookies)
	http.HandleFunc("/post-with-json", postWithJSON)
	http.HandleFunc("/postWithFormParams", postWithFormParams)
	http.HandleFunc("/put", put)
	http.HandleFunc("/delete", delete)

	err := http.ListenAndServe(":8091", nil)
	if err != nil {
		log.Fatal("Listen And Server:", err)
	}
}
func getWithQuery(w http.ResponseWriter, r *http.Request) {
	q := r.URL.RawQuery //  请求结果：query:q=golang&random=12345&scope=project
	//	fmt.Printf("%#+v\n",r.URL.Query())   //  返回  url.Values{"q":[]string{"golang"}, "random":[]string{"12345"}, "scope":[]string{"project"}}
	fmt.Fprintf(w, "query:%s", q)
}
func getTimeout(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(1) * time.Second)
	fmt.Fprintf(w, "http get timeout")
}

func postWithCookies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "need post")
		return
	}

	cookies, _ := json.Marshal(r.Cookies())
	w.Write(cookies)
	//fmt.Fprintf(w, "cookies:%s", cookies)
}

func postWithFormParams(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "need post")
		return
	}

	r.ParseForm()
	fmt.Printf("数据解析:%#+v\n", r.PostForm) //  数据解析:url.Values{"q":[]string{"golang"}, "random":[]string{"12345"}, "scope":[]string{"project"}}

	params, _ := json.Marshal(r.Form)

	fmt.Fprintf(w, "form params:%s", params)
}

func postWithJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "need post")
		return
	}

	json, _ := ioutil.ReadAll(r.Body)

	fmt.Fprintf(w, "json:%s", json)
}

func put(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		fmt.Fprintf(w, "need put")
		return
	}

	fmt.Fprintf(w, "http put")
}

func delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		fmt.Fprintf(w, "need delete")
		return
	}

	fmt.Fprintf(w, "http delete")
}
