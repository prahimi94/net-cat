# net-cat

## Description
This project consists on recreating the NetCat in a Server-Client Architecture that can run in a server mode on a specified port listening for incoming connections, and it can be used in client mode, trying to connect to a specified port and transmitting information to the server.

NetCat, nc system command, is a command-line utility that reads and writes data across network connections using TCP or UDP. It is used for anything involving TCP, UDP, or UNIX-domain sockets, it is able to open TCP connections, send UDP packages, listen on arbitrary TCP and UDP ports and many more.

To see more information about NetCat inspect the manual man nc.

Your project must work in a similar way that the original NetCat works, in other words, you must create a group chat. The project must have the following features :

TCP connection between server and multiple clients (relation of 1 to many).
A name requirement to the client.
Control connections quantity.
Clients must be able to send messages to the chat.
Do not broadcast EMPTY messages from a client.
Messages sent, must be identified by the time that was sent and the user name of who sent the message, example : [2020-01-20 15:48:41][client.name]:[client.message]
If a Client joins the chat, all the previous messages sent to the chat must be uploaded to the new Client.
If a Client connects to the server, the rest of the Clients must be informed by the server that the Client joined the group.
If a Client exits the chat, the rest of the Clients must be informed by the server that the Client left.
All Clients must receive the messages sent by other Clients.
If a Client leaves the chat, the rest of the Clients must not disconnect.
If there is no port specified, then set as default the port 8989. Otherwise, program must respond with usage message: [USAGE]: ./TCPChat $port

---

## Authors
- **[Parisa Rahimi Darabad]**
- **[Fatemeh Kheirkhah]**

---

## Usage
### How to Run
1. Clone this repository to your local machine:
   ```bash
   git clone https://01.gritlab.ax/git/prahimi/net-cat.git

2. Navigate to the project directory:
    ```bash
    cd net-cat

3. Build the project with this command:
    ```bash
    go build -o TCPChat

3. Start the TCP server: ($port is optional. if port is not specified the default port which is 8989 would be used)
    ```bash
    go run . $port
    or
    ./TCPChat $port

4. Clients can connect to server using this command:
    ```bash
    nc localhost $port

**Maximum count of clients are 10.**
    
## Project Structure and Implementation
Project runs on terminal

Backend: Include go file for handling server
