package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"text/template"
	"sync"
	"time"
	"math/rand"
	"encoding/json"
	"os"
)

var (
	gocalls chan int
	
	connections = struct {
		sync.RWMutex
		m map[*websocket.Conn]string
	}{m: make(map[*websocket.Conn]string)}
	
	upgrader = websocket.Upgrader {
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
	}
)

type Message struct {
	Text string
	Color string
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	
	connections.Lock()
	connections.m[conn] = getRandomColor()
	connections.Unlock()
	
	log.Println("Succesfully upgraded connection")
	
	for {
		_, msg, err := conn.ReadMessage()
		
		if err != nil {
			connections.Lock()
			delete(connections.m, conn)
			connections.Unlock()
			conn.Close()
			return
		}
		
		if string(msg) != "Keep connection alive!" {
			go sendAll(conn, msg)
		}
	}
}

func sendAll(sender *websocket.Conn, msg []byte) {
	calls := <- gocalls
	gocalls <- calls + 1
	log.Println("# of go calls:", calls)
	
	for conn := range connections.m {
		
		message, _ := json.Marshal(&Message{ Text: string(msg), Color: connections.m[sender] })
		
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			connections.Lock()
			delete(connections.m, conn)
			connections.Unlock()
			conn.Close()
		}
	}
}

func getRandomColor() string{
	
    letters := "0123456789ABCDEF"
    color := "#"
    
    rand.Seed(time.Now().UnixNano())
    
    for i := 0; i < 6; i++ {
    	num := rand.Intn(16)
    	color += letters[num: num + 1]
    }
      
    return color
}

func main() {
	
	port := GetPort()
	
	gocalls = make(chan int, 100)
	gocalls <- 0
	
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/css/", serveStatic)
	http.HandleFunc("/images/", serveStatic)
	http.HandleFunc("/js/", serveStatic)
	
	log.Printf("Running on port" + port)
	
	
	err := http.ListenAndServe(port, nil)
	log.Println(err.Error())
	
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	template.Must(template.ParseFiles("html/client.html")).Execute(w, nil)
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	template.Must(template.ParseFiles(r.URL.Path[1:])).Execute(w, nil)
}

func GetPort() string {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
        log.Println("[-] No PORT environment variable detected. Setting to ", port)
    }
    return ":" + port
}