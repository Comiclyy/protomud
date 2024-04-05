package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
)

// Movement represents the movement data received from the client
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

var (
	clients     = make(map[string]string) // Map to store client ID and IP address
	clientsLock sync.Mutex                // Mutex for concurrent access to clients map
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Receive client's movement data
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	// Decode JSON data
	var movement Movement
	err = json.Unmarshal(buffer[:n], &movement)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// Get client's IP address
	clientAddr := conn.RemoteAddr().String()

	// Normalize client's IP address
	clientAddr = strings.Split(clientAddr, ":")[0]

	// Add client to map
	clientID := movement.ClientID
	clientsLock.Lock()
	clients[clientID] = clientAddr
	numClients := len(clients)
	clientsLock.Unlock()

	fmt.Printf("Client connected [%s]: %s | Amount of users connected: [%d]\n", clientID, clientAddr, numClients)

	// Read and update player's coordinates
	playerData, err := ioutil.ReadFile(fmt.Sprintf("/workspaces/protomud/userdata/%s.json", clientID))
	if err != nil {
		if os.IsNotExist(err) {
			// Create player data file if it doesn't exist
			player := Player{X: 0, Y: 0, Z: 0}
			playerData, _ := json.MarshalIndent(player, "", " ")
			err := ioutil.WriteFile(fmt.Sprintf("/workspaces/protomud/userdata/%s.json", clientID), playerData, 0644)
			if err != nil {
				fmt.Println("Error creating player data file:", err)
				return
			}
		} else {
			fmt.Println("Error reading player data file:", err)
			return
		}
	}

	var player Player
	err = json.Unmarshal(playerData, &player)
	if err != nil {
		fmt.Println("Error decoding player data JSON:", err)
		return
	}

	switch strings.ToUpper(movement.Direction) {
	case "W":
		player.Y++
	case "A":
		player.X--
	case "S":
		player.Y--
	case "D":
		player.X++
	}

	// Print player's updated coordinates
	fmt.Printf("[%s] X: %d, Y: %d, Z: %d\n", clientID, player.X, player.Y, player.Z)

	// Update player data file
	playerData, _ = json.MarshalIndent(player, "", " ")
	err = ioutil.WriteFile(fmt.Sprintf("/workspaces/protomud/userdata/%s.json", clientID), playerData, 0644)
	if err != nil {
		fmt.Println("Error updating player data file:", err)
		return
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			return
		}
		go handleConnection(conn)
	}
}
