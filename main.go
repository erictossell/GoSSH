package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	Users      map[string]string `json:"users"`
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
func runSSHCommand(server, user, command, sshOptions string, results chan<- ServerResult) {
	// Construct the command string for debugging
	//var commandString string
	//if runtime.GOOS == "windows" {
	//		commandString = fmt.Sprintf(`ssh %s %s@%s "%s"`, sshOptions, user, server, command)
	//	} else {
	//		commandString = fmt.Sprintf(`ssh %s %s@%s %s`, sshOptions, user, server, command)
	//	}

	startTime := time.Now()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", fmt.Sprintf(`ssh %s %s@%s "%s"`, sshOptions, user, server, command))
	} else {
		cmd = exec.Command("ssh", sshOptions, fmt.Sprintf("%s@%s", user, server), command)
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	configDir := filepath.Join(homeDir, ".config", "GoSSH")
	configFile := filepath.Join(configDir, "configuration.json")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatalf("Error creating config directory: %v", err)
		}
		exampleConfig := Config{
			Servers:    []string{"example.server.com"},
			SSHOptions: map[string]string{"example.server.com": "-p 22"},
			Users:      map[string]string{"example.server.com": "example"},
			// Add other configuration fields here",
		}
		exampleConfigBytes, err := json.MarshalIndent(exampleConfig, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling example config: %v", err)
		}
		if err := os.WriteFile(configFile, exampleConfigBytes, 0644); err != nil {
			log.Fatalf("Error writing example config: %v", err)
		}
		fmt.Printf("Example configuration file created at %s. Please edit it and run the program again.\n", configFile)
		os.Exit(1)
	}

	config, err := readConfig(configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: GoSSH <command1> <command2> ...")
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
		user := config.Users[server] // Retrieve the user for the server
		for _, command := range commands {
			wg.Add(1)
			go func(server, user, command, sshOptions string) {
				defer wg.Done()
				runSSHCommand(server, user, command, sshOptions, results)
			}(server, user, command, sshOptions)
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
