package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	/// obtaining configurations from INI
	// gets INI file location from registry
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\INI_FILE`, registry.QUERY_VALUE)
	if err != nil {
		fmt.Printf("Fail to get INI file location: %v", err)
		os.Exit(1)
	}
	// A defer statement defers the execution of a function until the surrounding function returns.
	defer k.Close()
	INI_DIR, _, _ := k.GetStringValue("Base Directory")
	INI_DIR = fmt.Sprintf("%s/conf.ini", INI_DIR)
	/// end

	/// loading INI to memory
	cfg, err := ini.Load(INI_DIR)
	if err != nil {
		fmt.Printf("Fail to read INI: %v", err)
		os.Exit(1)
	}
	// gets info from INI
	TransitID := cfg.Section("CONFIG").Key("ID").String()
	DomainName := cfg.Section("CONFIG").Key("DomainName").String()
	Authorization := cfg.Section("CONFIG").Key("Authorization").String()
	Authorization = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(Authorization)))
	// sets TLS CA certificate
	caCertPath := fmt.Sprintf("%s/cacert.pem", INI_DIR)
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		fmt.Printf("Fail to read TLS CA certificate: %v", err)
		os.Exit(1)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	/// end

	/// HTTPS GET
	// creates a TLS client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
	// sets URL
	URL := fmt.Sprintf("https://%s", DomainName)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		fmt.Printf("Fail to create HTTPS GET: %v", err)
		os.Exit(2)
	}
	// sets HTTP headers
	req.Header.Set("Authorization", Authorization)
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("ID", ID)
	// makes HTTPS request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Fail to make HTTPS request: %v", err)
		os.Exit(3)
	}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Fail to get HTTPS response: %v", err)
		os.Exit(4)
	}
	// reads HTTP status code
	defer resp.Body.Close()
	status := strings.Split(resp.Status, " ")[1]
	// shows GET result
	if status == "OK" {
		for _, item := range contents {
			fmt.Println(item)
		}
	} else {
		os.Exit(5)
	}
	/// end
}
/// end of file
