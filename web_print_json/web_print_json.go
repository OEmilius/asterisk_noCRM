// packet_list_window project packet_list_window.go
package web_print_json

//package web

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

var WebHost_port string = ":8088"

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var client_number int = 0

var Clients = make(map[int]WebClient, 10)
var Cmd_ch = make(chan Cmd)

type Cmd struct {
	Id  int
	Msg string
}

type WebClient struct {
	Id int
	//Cmd_ch chan string
	Ans_ch chan string
}

func NewWebClient() (WebClient, int) {
	id := client_number + 1
	client_number++
	wc := WebClient{
		Id: id,
		//Cmd_ch: make(chan string),
		Ans_ch: make(chan string, 500)}
	Clients[id] = wc
	return wc, id
}

func (wc *WebClient) Destroy() {
	//close(wc.Cmd_ch)
	close(wc.Ans_ch)
	delete(Clients, wc.Id)
}

//в этот канал будут приходить данные для отображения на странице
//var Clients_ch = make(map[int]chan string)

//этот канал в нем будут приходить команды с адресом канала для ответа
//var Cmd_ch = make(map[int]chan Cmd)

//type Cmd struct {
//	Msg     string      //команда
//	Answ_ch chan string //куда отвечать
//}

//то что пришло в websocket от клиента
//var Filter string = "000000000000"

var (
	packet_list_Templ = template.Must(template.ParseFiles("index.html"))
	upgrader          = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func reader(ws *websocket.Conn, client WebClient) {
	//fmt.Println("reader started")
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		mt, msg, err := ws.ReadMessage()
		//fmt.Println("new message ", mt, msg)
		//сюда добавить прием сообщений
		if err != nil {
			break
		}
		switch mt {
		case websocket.TextMessage: //пришло текстовое сообщение mt = 1
			fmt.Println("from web socket", client.Id, string(msg))
			Cmd_ch <- Cmd{Id: client.Id, Msg: string(msg)}
			//fmt.Println("from client ch =", client_number, "rcv", string(msg))
		case websocket.PingMessage: //9
			fmt.Println("ping message")
		case websocket.PongMessage: //10
			fmt.Println("pong message")
		case websocket.CloseMessage: //8
			fmt.Println("close message")
		}

	}
	fmt.Println("reader finished")
	Cmd_ch <- Cmd{Id: client.Id, Msg: "close_ws:"}
}

//func Read_Cmd_ch() {
//	for {
//		for _, ch := range Cmd_ch {
//			c := <-ch
//			fmt.Println("msg=", c.Msg, "answ_ch=", c.Answ_ch)
//			c.Answ_ch <- "получил от тебя команду " + c.Msg
//		}
//	}
//}

//func Read_all_commands() {
//	for {
//		for k, ch := range Command_ch {
//			if cmd, ok := <-ch; ok {
//				fmt.Println("from ", k, "cmd=", cmd)
//				break
//			}
//		}
//	}
//}

func my_writer(ws *websocket.Conn, client WebClient) {
	fmt.Println("my_writer started")
	pingTicker := time.NewTicker(pingPeriod)
	//	dataTicket := time.NewTicker(1 * time.Second)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	go func() {
		for {
			<-pingTicker.C
			fmt.Println("ping from pingTicker")
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}

		}
	}()
	for {
		//		select {
		//		//case <-dataTicket.C:
		//		//	//data_chan <- time.Now().String() + "\r\n"
		//		//	row := "<td>" + strconv.Itoa(number) + "</td><td>" + time.Now().String() + "</td>"
		//		//	data_chan <- row
		//		//	number += 1
		//		case my_data := <-Clients_ch[client_number]:
		//			ws.SetWriteDeadline(time.Now().Add(writeWait))
		//			if err := ws.WriteMessage(websocket.TextMessage, []byte(my_data)); err != nil {
		//				return
		//			}
		//		case <-pingTicker.C:
		//			ws.SetWriteDeadline(time.Now().Add(writeWait))
		//			fmt.Println("ping from pingTicker")
		//			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		//				return
		//			}

		//		}
		my_data := <-client.Ans_ch
		//Attention !!!
		//ws.SetWriteDeadline(time.Now().Add(writeWait))
		if err := ws.WriteMessage(websocket.TextMessage, []byte(my_data)); err != nil {
			return
		}
	}
	fmt.Println("my_writer finished")
}
func serveWs(w http.ResponseWriter, r *http.Request) {
	//client_number += 1
	wclient, id := NewWebClient()
	fmt.Println("Запустился serveWs", id)
	Cmd_ch <- Cmd{Id: id, Msg: "newWebClient:"}
	//Clients_ch[client_number] = make(chan string, 500)
	//Cmd_ch[client_number] = make(chan Cmd)
	//Command_ch[client_number] = make(chan string, 10)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			fmt.Println(err)
		}
		//
		return
	}
	//go writer(ws, lastMod)
	go my_writer(ws, wclient)
	reader(ws, wclient)
	//close(Clients_ch[client_number])
	//delete(Clients_ch, client_number)
	//delete(Cmd_ch, client_number)
	wclient.Destroy()
	fmt.Println("servWS finished")
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
	}{r.Host, ""}
	packet_list_Templ.Execute(w, &v)
}

/*
func main() {
	go Start()
	_ = exec.Command("cmd", "/c", "start", "http://localhost:8081/").Start()
	//go Read_all_commands()
	go Read_Cmd_ch()
	time.Sleep(5 * time.Second)
	for _, ch := range Clients_ch {
		ch <- `{"type":"message", "id":"tr7", "calling":"78998", "called":"495", "state":"setup", "packet":"INVITE"}`
		ch <- `{"type":"message", "id":"tr7", "calling":"78998", "called":"495", "state":"setup", "packet":"100"}`
		//ch <- `{"type":"message", "text":"200 ok", "id":1}`
		//ch <- `{"type":"message", "text":"invite ok", "id":1}`
	}
	Clients_ch[1] <- `{"type":"message", "text":"invite ok", "id":1}`
	//Clients_ch[1] <- "ssssssssssssssssss"
	time.Sleep(10 * time.Second)
	//Clients_ch[2] <- "2222222222222222"
	fmt.Scanln()
}*/

func Start() {
	fmt.Println("http.ListenAndServe:8088")
	//go stop_web_server() //что бы сервер сам остановился
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	//Data_chan <- "\r\nsadfasdf\r\n"
	if err := http.ListenAndServe(WebHost_port, nil); err != nil {
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
	_ = exec.Command("cmd", "/c", "start", "http://localhost:8088/").Start()
}
