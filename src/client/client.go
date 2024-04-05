package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Movement represents the movement data sent to the server
type Movement struct {
	ClientID  string `json:"client_id"`
	Direction string `json:"direction"`
}

// Player represents the player's coordinates
type Player struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

func generateClientID() string {
	rand.Seed(time.Now().UnixNano())
	clientID := rand.Intn(9000) + 100 // Generate a random 4-digit number
	return fmt.Sprintf("%04d", clientID)
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	clientID := generateClientID()
	fmt.Printf("Connected to server with Client ID: [%s]\n", clientID)

	// Create player data JSON file
	player := Player{X: 0, Y: 0, Z: 0}
	playerData, _ := json.MarshalIndent(player, "", " ")
	err = os.WriteFile(fmt.Sprintf("/workspaces/protomud/userdata/%s.json", clientID), playerData, 0644)
	if err != nil {
		fmt.Println("Error creating player data file:", err)
		return
	}

	// Listen for signals to gracefully close the connection
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Goroutine to read signals and close connection gracefully
	go func() {
		<-sigCh
		fmt.Println("\nClosing connection...")
		conn.Close()
		os.Exit(0)
	}()

	for {
		var input string
		fmt.Print("Enter direction (W/A/S/D): ")
		fmt.Scanln(&input)

		// Normalize direction to uppercase
		input = strings.ToUpper(input)

		// Construct JSON data based on user input
		movement := Movement{ClientID: clientID, Direction: input}
		jsonData, err := json.Marshal(movement)
		if err != nil {
			fmt.Println("Error encoding JSON:", err)
			continue
		}

		// Send JSON data to server
		_, err = conn.Write(jsonData)
		if err != nil {
			fmt.Println("Error sending:", err)
			fmt.Println("Disconnected from server.")
			return
		}
		fmt.Println("Sent movement data to server:", string(jsonData))
	}
}
