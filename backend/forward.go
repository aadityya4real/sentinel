package main

import (
"bufio"
"fmt"
"io"
"net/http"
"os"
"os/exec"
"path/filepath"
"strings"
"time"
)

func main() {
	tempFile := filepath.Join(os.TempDir(), "agent_forward_input.json")

	cmd := exec.Command("W:\\sentinel\\backend\\agent.exe")
	outFile, err := os.Create(tempFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create temp file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Start()
	defer cmd.Process.Kill()
	fmt.Println("Agent started, PID:", cmd.Process.Pid)

	lastCount := 0
	for {
		f, err := os.Open(tempFile)
		if err != nil { time.Sleep(1 * time.Second); continue }
		allLines := []string{}
		s := bufio.NewScanner(f)
		for s.Scan() { allLines = append(allLines, s.Text()) }
		f.Close()
		newCount := len(allLines) - lastCount
		if newCount > 0 {
			for i := lastCount; i < len(allLines); i++ {
				line := strings.TrimSpace(allLines[i])
				if line == "" { continue }
				resp, err := http.Post("http://localhost:8080/api/v1/metrics", "application/json", strings.NewReader(line))
				if err != nil { fmt.Printf("POST error: %v\n", err); continue }
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			lastCount = len(allLines)
		}
		time.Sleep(2 * time.Second)
	}
}