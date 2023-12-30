package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"sync"
	"time"
)

type ServerResult struct {
	Server   string
	Output   string
	Error    error
	Duration float64
}

type Config struct {
	Servers    []string          `json:"servers"`
	SSHOptions map[string]string `json:"ssh_options"` // key: server, value: options
	// Add other configuration fields here
}

func readConfig(filePath string) (*Config, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
func runSSHCommand(server, command, sshOptions string, results chan<- ServerResult) {
	startTime := time.Now()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", fmt.Sprintf(`ssh %s eriim@%s "%s"`, sshOptions, server, command))
	} else {
		cmd = exec.Command("ssh", sshOptions, fmt.Sprintf("eriim@%s", server), command)
	}

	outputBytes, err := cmd.CombinedOutput()
	duration := time.Since(startTime).Seconds()
	result := ServerResult{
		Server:   server,
		Output:   string(outputBytes),
		Error:    err,
		Duration: duration,
	}
	results <- result
}

func main() {
	// Read config
	config, err := readConfig("configuration.json")
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: mysshcommand <command1> <command2> ...")
		os.Exit(1)
	}

	// Command-line arguments
	commands := os.Args[1:]

	servers := []string{"192.168.2.195", "192.168.2.196", "192.168.2.197"}

	// Log file
	logFile, err := os.OpenFile("deploy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	results := make(chan ServerResult, len(config.Servers)*len(commands))
	var wg sync.WaitGroup

	for _, server := range config.Servers {
		sshOptions := config.SSHOptions[server]
		for _, command := range commands {
			wg.Add(1)
			go func(server, command, sshOptions string) {
				defer wg.Done()
				runSSHCommand(server, command, sshOptions, results)
			}(server, command, sshOptions)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var serverResults []ServerResult
	for result := range results {
		serverResults = append(serverResults, result)
	}

	// Sorting results based on the order of servers
	sort.Slice(serverResults, func(i, j int) bool {
		for _, server := range servers {
			if serverResults[i].Server == server {
				return true
			} else if serverResults[j].Server == server {
				return false
			}
		}
		return false
	})

	// Output results
	for _, result := range serverResults {
		if result.Error != nil {
			log.Printf("Error from %s: %v (Duration: %.2fs)\n", result.Server, result.Error, result.Duration)
			fmt.Printf("Error from %s: %v (Duration: %.2fs)\n", result.Server, result.Error, result.Duration)
		} else {
			log.Printf("Output from %s:\n%s\n(Duration: %.2fs)\n", result.Server, result.Output, result.Duration)
			fmt.Printf("Output from %s:\n%s\n(Duration: %.2fs)\n", result.Server, result.Output, result.Duration)
		}
	}
	fmt.Println("Execution completed on all servers.")
}
