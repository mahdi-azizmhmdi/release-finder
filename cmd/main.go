package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"release_finder/config"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/mpvl/unique"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var VersionHistory = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "New_Version_Of_Services",
		Help: " The Latest Version Of Devops Services",
	},
	[]string{"version", "service"},
)
var c = cache.New(1*time.Hour, 2*time.Hour)
var Lasturl []string

func main() {
	c := config.Load()
	prometheus.MustRegister(VersionHistory)
	ticker := time.NewTicker(1 * time.Hour)
	// defer ticker.Stop()
	go func() {
		for ; ; <-ticker.C {
			prometheus.Unregister(VersionHistory)
			VersionHistory = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "New_Version_Of_Services",
					Help: " The Latest Version Of Devops Services",
				},
				[]string{"version", "service"},
			)
			prometheus.MustRegister(VersionHistory)
			for _, service := range c.Services {
				fmt.Printf("\nService: %s\n", service.Service)
				data := MakeRequest(service.URL)
				version := Parser(data, service.URL)
				fmt.Println(version)
				VersionHistory.DeleteLabelValues(service.Service)
				VersionHistory.WithLabelValues(version, service.Service).Set(1)
				Lasturl = append(Lasturl, service.URL)

			}
			go HealthCheck(Lasturl)
		}

	}()

	StartMetricsServer()

}

func StartMetricsServer() {

	router := mux.NewRouter()
	router.Path("/metrics").Handler(promhttp.Handler())
	fmt.Printf("\nServing Prometheus metrics on port 9000")
	log.Fatal(http.ListenAndServe(":9000", router))
}

func Parser(data []byte, url string) string {
	var jsonRes interface{}
	_ = json.Unmarshal(data, &jsonRes)

	if strings.Contains(url, "tags") {
		if arr, ok := jsonRes.([]interface{}); ok && len(arr) > 0 {
			if tag, ok := arr[0].(map[string]interface{}); ok {
				if name, ok := tag["name"].(string); ok {
					return name
				}
			}
		}
	} else if strings.Contains(url, "releases") {
		if m, ok := jsonRes.(map[string]interface{}); ok {
			if tagName, ok := m["tag_name"].(string); ok {
				return tagName
			}
		}
	} else {
		fmt.Println("Unknown URL pattern:", url)
		return ""
	}
	return ""
}

func MakeRequest(Url string) []byte {

	if cachedResponse, found := c.Get(Url); found {
		return cachedResponse.([]byte)
	}
	client := &http.Client{}
	req, _ := http.NewRequest("GET", Url, nil)
	req.Header.Set("User-Agent", "release-finder-App (https://api.github.com)")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, resp.Body)
	}
	if err != nil {
		log.Fatal(err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	c.Set(Url, respBody, cache.DefaultExpiration)
	return respBody
}

func HealthCheck(urls []string) {
	var uniqueurls []string
	for _, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}
		uniqueurl := parsedURL.Hostname()
		uniqueurls = append(uniqueurls, uniqueurl)
	}
	unique.Strings(&uniqueurls)
	health := healthcheck.NewHandler()
	for _, domain := range uniqueurls {
		// fmt.Println(reflect.TypeOf(domain))
		health.AddLivenessCheck("upstream-dep-check", healthcheck.DNSResolveCheck(domain, 60*time.Second))
	}
	http.ListenAndServe("0.0.0.0:9001", health)

}
