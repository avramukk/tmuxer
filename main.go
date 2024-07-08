package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "tmuxer",
	Short: "A CLI tool to manage preconfigurable Tmux sessions",
	Run: func(cmd *cobra.Command, args []string) {
		displayConfig()
		startSessions()
		attachToFirstSession()
	},
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s\n", err)
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type Pane struct {
	Command string `mapstructure:"command"`
}

type Window struct {
	Name  string `mapstructure:"name"`
	Panes []Pane `mapstructure:"panes"`
}

type Session struct {
	Name    string   `mapstructure:"name"`
	Windows []Window `mapstructure:"windows"`
}

type Config struct {
	Sessions []Session `mapstructure:"sessions"`
}

func displayConfig() {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Unable to decode into struct, %v\n", err)
		return
	}

	fmt.Println("Tmux Configuration:")
	for _, session := range config.Sessions {
		fmt.Printf("Session: %s\n", session.Name)
		for _, window := range session.Windows {
			fmt.Printf("  Window: %s\n", window.Name)
			for i, pane := range window.Panes {
				fmt.Printf("    Pane %d: %s\n", i+1, pane.Command)
			}
		}
	}
	fmt.Println()
}

func startSessions() {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Unable to decode into struct, %v\n", err)
		return
	}

	for _, session := range config.Sessions {
		startSession(session)
	}
}

func startSession(session Session) {
	execCommand(fmt.Sprintf("tmux new-session -d -s %s", session.Name))

	for _, window := range session.Windows {
		startWindow(session.Name, window)
	}
}

func startWindow(sessionName string, window Window) {
    execCommand(fmt.Sprintf("tmux new-window -t %s -n %s", sessionName, window.Name))

    for i, pane := range window.Panes {
        if i == 0 {
            execCommand(fmt.Sprintf("tmux split-window -t %s:%s -v '%s'", sessionName, window.Name, pane.Command))
        } else {
            execCommand(fmt.Sprintf("tmux split-window -t %s:%s -v -l 10 '%s'", sessionName, window.Name, pane.Command))
        }
    }

    execCommand(fmt.Sprintf("tmux select-layout -t %s:%s tiled", sessionName, window.Name))
}

func execCommand(command string) {
	if err := exec.Command("bash", "-c", command).Run(); err != nil { // -c 
		fmt.Printf("Error executing command: %s\n", err)
	}
}

func attachToFirstSession() {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Unable to decode into struct, %v\n", err)
		return
	}

	if len(config.Sessions) > 0 {
		firstSession := config.Sessions[0].Name
		fmt.Printf("Attaching to the first session: %s\n", firstSession)
		cmd := exec.Command("tmux", "has-session", "-t", firstSession)
		if err := cmd.Run(); err != nil {
			// If session doesn't exist, create it
			fmt.Printf("Session %s doesn't exist, creating it...\n", firstSession)
			execCommand(fmt.Sprintf("tmux new-session -d -s %s", firstSession))
		}

		// Attempt to attach to the session
		cmd = exec.Command("tmux", "attach-session", "-t", firstSession)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error attaching to session: %s\n", err)
		}
	}
}
