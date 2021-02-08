package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type counters struct {
	sync.Mutex
	data  string
	view  int
	click int
}
type json_cont struct {
	Time  string
	Views int
	Click int
}

var (
	c           = counters{}
	Max_REQUEST = 10 //for rate limiting
	obj         = json_cont{
		Time: "", Views: 0, Click: 0,
	}
	m         = make(map[string]int)
	ip_to_req = make(map[string]int)
	content   = []string{"sports", "entertainment", "business", "education"}
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	data := content[rand.Intn(len(content))]

	c.Lock()
	c.view++
	c.Unlock()

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	// simulate random click call
	if rand.Intn(100) < 50 {
		processClick(data)
	}
}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {
	c.Lock()
	c.click++
	c.data = data
	c.Unlock()
	m["views"] = c.view
	m["click"] = c.click

	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed() {
		w.WriteHeader(429)
		return
	}
	fmt.Fprint(w, "keep reloading, 10 requests not exceeded")
}
func start_writing1() {
	for {
		time.Sleep(5 * time.Second)
		go write()
	}
}

func start_writing2() {
	for {
		<-time.After(5 * time.Second)
		go write()
	}
}

func write() {
	filename := "mock_store.json"
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Print("error")
	}
	data := []json_cont{}
	json.Unmarshal(file, &data)
	time := time.Now()

	obj := &json_cont{
		Time:  c.data + ":" + time.String(),
		Views: m["views"],
		Click: m["click"],
	}
	data = append(data, *obj)
	test, err := json.MarshalIndent(data, "", "")
	if err != nil {
		fmt.Print("error")
	}
	ioutil.WriteFile(filename, test, 0644)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value, ok := ip_to_req[GetIP(r)]
		if ok {
			if value > Max_REQUEST {
				fmt.Fprint(w, "Too many request, please try later")
				return
			}
			ip_to_req[GetIP(r)]++
		} else {
			ip_to_req[GetIP(r)] = 1
		}
		next.ServeHTTP(w, r)
	})
}
func GetIP(r *http.Request) string {
	forward := r.Header.Get("X-FORWARDED-FOR")
	if forward != "" {
		return forward
	}
	return r.RemoteAddr
}

func isAllowed() bool {
	return true
}

func uploadCounters() error {
	return nil
}

func main() {
	go start_writing1()
	go start_writing2()
	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)
	s_h := http.HandlerFunc(statsHandler)
	http.Handle("/stats/", middleware(s_h))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
