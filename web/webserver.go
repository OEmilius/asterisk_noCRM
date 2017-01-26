// packet_list_window project packet_list_window.go
package web

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

//в этот канал будут приходить данные для отображения на странице
var Data_chan = make(chan string, 10)

//то что пришло в websocket от клиента
var Filter string = "000000000000"

var (
	//packet_list_Templ = template.Must(template.New("").Parse(packet_list))
	packet_list_Templ = template.Must(template.ParseFiles("web/packet_window.html"))
	upgrader          = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		mt, msg, err := ws.ReadMessage()
		//сюда добавить прием сообщений
		if err != nil {
			break
		}
		switch mt {
		case websocket.TextMessage: //пришло текстовое сообщение mt = 1
			fmt.Println(string(msg))
			Filter = string(msg) //в главной программе применим это выражение для фильтра
		case websocket.PingMessage: //9
			fmt.Println("ping message")
		case websocket.PongMessage: //10
			fmt.Println("pong message")
		case websocket.CloseMessage: //8
			fmt.Println("close message")
		}

	}
}

func my_writer(ws *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	//	dataTicket := time.NewTicker(1 * time.Second)

	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		//case <-dataTicket.C:
		//	//data_chan <- time.Now().String() + "\r\n"
		//	row := "<td>" + strconv.Itoa(number) + "</td><td>" + time.Now().String() + "</td>"
		//	data_chan <- row
		//	number += 1
		case my_data := <-Data_chan:
			//fmt.Println("что то пришло в канал из вне на отправку web браузеру")
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, []byte(my_data)); err != nil {
				return
			}
		}
	}
}
func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			fmt.Println(err)
		}
		//
		return
	}
	//go writer(ws, lastMod)
	go my_writer(ws)
	reader(ws)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var v = struct {
		Host string
		Data string
	}{
		r.Host,
		"",
		//"some_data",
	}
	packet_list_Templ.Execute(w, &v)
}

//func main() {
func Start() {
	fmt.Println("http.ListenAndServe:8082")
	//go stop_web_server() //что бы сервер сам остановился
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	//Data_chan <- "\r\nsadfasdf\r\n"
	if err := http.ListenAndServe(":8082", nil); err != nil {
		fmt.Println(err)
	}
}

func stop_web_server() {
	st := time.NewTimer(25 * time.Second)
	if _, ok := <-st.C; ok {
		fmt.Println("time to stop server", time.Now())
		os.Exit(0)
	}
}

func Start_and_open() {
	go Start()
	_ = exec.Command("cmd", "/c", "start", "http://localhost:8082/").Start()
}
