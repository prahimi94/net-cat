package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type connectionCount struct {
	number int
	mutex  sync.Mutex
}

type client struct {
	index          int
	number         int
	name           string
	messageChannel chan string
	occupied       bool
}

var (
	WarningLog *log.Logger
	InfoLog    *log.Logger
	ErrorLog   *log.Logger
)

var clients [10]client

var historyFileName string

func init() {
	currentTime := time.Now().Format("20060102150405")

	errHistoryDir := createDirIfNotExists("historyFiles")
	if errHistoryDir != nil {
		fmt.Printf("Error creating directory: %v\n", errHistoryDir)
		return
	}

	errLogFiles := createDirIfNotExists("logFiles")
	if errLogFiles != nil {
		fmt.Printf("Error creating directory: %v\n", errLogFiles)
		return
	}

	historyFileName = "historyFiles/history-" + currentTime + ".txt"

	logFile, err := os.OpenFile("logFiles/log-"+currentTime+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLog = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLog = log.New(logFile, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	port, err := processInput()
	checkError(err)

	establishConnection(port)
}

func processInput() (string, error) {
	port := "8989"

	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		WarningLog.Println("[USAGE]: ./TCPChat $port")
		os.Exit(1)
	} else if len(os.Args) == 2 {
		port = os.Args[1]
	}
	return port, nil
}

func establishConnection(port string) {
	connectionCount := connectionCount{}

	// Listen for incoming connections on port
	ln, err := net.Listen("tcp", ":"+port)
	checkError(err)
	defer ln.Close()

	fmt.Println("Server started on port " + port)
	InfoLog.Println("Server started on port " + port)

	go readServerMessages()

	for i := 0; i < 10; i++ {
		clients[i].index = i
		clients[i].number = i + 1
		clients[i].messageChannel = make(chan string)
	}

	connectionCount.number = 0
	wg := sync.WaitGroup{}

	// Accept incoming connections and handle them
	for {
		conn, err := ln.Accept()
		checkError(err)

		if connectionCount.number >= 10 {
			conn.Write([]byte("Server is full! Please try again later\n"))

			errorMessage := "One client rejected because the server is full"
			fmt.Println(errorMessage)
			WarningLog.Println(errorMessage)

			conn.Close()
			continue
		}
		wg.Add(1)

		var passingClientIndex int
		for i, client := range clients {
			if !client.occupied {
				clients[i].occupied = true
				passingClientIndex = i
				break
			}
		}

		go handleConnection(conn, passingClientIndex, &clients, &connectionCount, &wg)

		go broadcastMessage(conn, passingClientIndex, &clients)
		wg.Wait()
	}
}

func readServerMessages() {
	// Create a buffered reader to read input from the server console
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read the message typed by the server operator
		serverMessage, _ := reader.ReadString('\n')

		if strings.HasPrefix(serverMessage, "--clientNames") {
			for _, client := range clients {
				if client.name != "" {
					fmt.Println(client.name)
				}
			}
		}
	}
}

func handleConnection(conn net.Conn, passingClientIndex int, clients *[10]client, connectionCount *connectionCount, wg *sync.WaitGroup) {
	defer conn.Close() // Ensure connection is closed only when the client disconnects.
	connectionCount.mutex.Lock()
	// connectionCountNumber := connectionCount.number
	connectionCount.number++
	connectionCount.mutex.Unlock()

	wg.Done()

	currentTime := time.Now().Format("2006-01-02 15:04:05")

	welcomeMessage, err := readFile("./welcome.txt", "complete")
	checkError(err)

	conn.Write([]byte(welcomeMessage))

	InfoLog.Println("Client connected:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	messageCounter := 1
	var clientName string
	for {
		// Read message from client
		message, err := reader.ReadString('\n') // Assuming messages are newline-terminated
		if err != nil {
			break // Exit the loop and close the connection
		}

		if messageCounter == 1 {
			status := true
			status, clientName = handleClientName(message, conn, clients, passingClientIndex, currentTime)
			if !status {
				continue
			}
		} else {
			status := true
			status, clientName = handleClientMessage(clientName, message, conn, clients, passingClientIndex, currentTime)
			if !status {
				continue
			}
		}
		messageCounter++
	}

	handleClientExit(clientName, clients, passingClientIndex, connectionCount, currentTime)
}

func broadcastMessage(conn net.Conn, passingClientIndex int, clients *[10]client) {
	for {
		for message := range clients[passingClientIndex].messageChannel {
			if clients[passingClientIndex].name == "" {
				continue
			}
			conn.Write([]byte(message))
		}
	}
}

func handleClientName(message string, conn net.Conn, clients *[10]client, passingClientIndex int, currentTime string) (bool, string) {
	if strings.TrimSpace(message) == "" {
		clearClientLastInput(conn)

		conn.Write([]byte("[ENTER YOUR NAME]: "))
		return false, ""
	}
	i := strings.LastIndex(message, "\n")
	clientName := message[:i] + strings.Replace(message[i:], "\n", "", 1)

	for _, client := range *clients {
		if client.name == clientName {
			// name is duplicated
			conn.Write([]byte("Rename failed due to duplicate name\n"))
			return false, ""
		}
	}

	clients[passingClientIndex].name = clientName

	// fmt.Printf("[%s]: %s has joined the chat...\n", currentTime, clientName)
	InfoLog.Printf("[%s]: %s has joined the chat...\n", currentTime, clientName)
	fillInChannnel(passingClientIndex, "notMe", *clients, fmt.Sprintf("%s has joined the chat...\n", clientName))
	writeInFile(historyFileName, fmt.Sprintf("%s has joined the chat...\n", clientName))

	// Call the function
	lines, err := readFile(historyFileName, "notLastLine")
	checkError(err)
	if len(lines) > 0 {
		lines += "\n"
	}

	conn.Write([]byte(lines))
	return true, clientName
}

func handleClientMessage(clientName string, message string, conn net.Conn, clients *[10]client, passingClientIndex int, currentTime string) (bool, string) {
	if strings.TrimSpace(message) == "" {
		return false, clientName
	} else if strings.HasPrefix(message, "--rename=") {
		clientRenameStatus, clientRename := handleRename(clientName, message, conn, clients, passingClientIndex, currentTime)
		if clientRenameStatus {
			clientName = clientRename
		}
		return false, clientName
	}

	clearClientLastInput(conn)

	messageText := message

	fillInChannnel(passingClientIndex, "all", *clients, fmt.Sprintf("[%s] [%s]: %s", currentTime, clientName, messageText))
	writeInFile(historyFileName, fmt.Sprintf("[%s] [%s]: %s", currentTime, clientName, messageText))

	return true, clientName
}

func handleRename(clientName string, message string, conn net.Conn, clients *[10]client, passingClientIndex int, currentTime string) (bool, string) {
	clientOldName := clientName
	clientName = strings.TrimPrefix(message, "--rename=")
	clientName = strings.TrimSuffix(clientName, "\n")

	for _, client := range *clients {
		if client.name == clientName {
			// name is duplicated
			conn.Write([]byte("Rename failed due to duplicate name\n"))
			return false, ""
		}
	}

	clients[passingClientIndex].name = clientName

	clearClientLastInput(conn)

	// fmt.Printf("[%s]: %s renamed their name to %s\n", currentTime, clientOldName, clientName)
	InfoLog.Printf("[%s]: %s renamed their name to %s\n", currentTime, clientOldName, clientName)
	fillInChannnel(passingClientIndex, "all", *clients, fmt.Sprintf("[%s]: %s renamed their name to %s\n", currentTime, clientOldName, clientName))
	writeInFile(historyFileName, fmt.Sprintf("[%s]: %s renamed their name to %s\n", currentTime, clientOldName, clientName))

	return true, clientName
}

func handleClientExit(clientName string, clients *[10]client, passingClientIndex int, connectionCount *connectionCount, currentTime string) {
	// fmt.Printf("[%s]: %s has left the chat...\n", currentTime, clientName)
	if clientName != "" {
		fillInChannnel(passingClientIndex, "all", *clients, fmt.Sprintf("%v has left our chat...\n", clientName))
		writeInFile(historyFileName, fmt.Sprintf("%v has left our chat...\n", clientName))
	}
	InfoLog.Printf("[%s]: %s has left the chat...\n", currentTime, clientName)

	clients[passingClientIndex].name = ""
	clients[passingClientIndex].messageChannel = make(chan string)
	clients[passingClientIndex].occupied = false

	connectionCount.mutex.Lock()
	connectionCount.number--
	connectionCount.mutex.Unlock()
}

func fillInChannnel(connectionNumber int, statusSend string, clients [10]client, message string) {
	for i, client := range clients {
		if statusSend == "notMe" && i == connectionNumber {
			continue
		} else {
			if client.messageChannel != nil {
				select {
				case client.messageChannel <- message:
					// Message sent successfully
				default:
					// Channel is full, message dropped (or handle as needed)
				}
			}
		}

	}
}

func clearClientLastInput(conn net.Conn) {
	// Clear the current line in the terminal
	conn.Write([]byte("\033[A\033[K"))
}

func createDirIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0755) // Use MkdirAll to create parent directories if needed
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}
	return nil
}

func readFile(fileName string, status string) (string, error) {
	// Open the file
	file, err := os.Open(fileName)
	checkError(err)

	defer file.Close()

	// Read lines using a scanner
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// Exclude the last line if there are any lines
	if len(lines) > 0 && status == "notLastLine" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n"), nil
}

func writeInFile(fileName string, message string) {
	// Open the file in append mode. Create the file if it doesn't exist.
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	checkError(err)

	defer file.Close()

	// Write the message to the file.
	_, err = file.WriteString(message)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		ErrorLog.Println("Error:", err)
		os.Exit(1)
	}
}
