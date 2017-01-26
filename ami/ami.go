package ami

import (
	"fmt"
	"net/textproto"
	"strconv"
)

type Manager struct {
	Address            string
	Conn               *textproto.Conn
	Events_chan        chan textproto.MIMEHeader
	Responce_chan      chan textproto.MIMEHeader
	Enents_string_chan chan string
}

var actionID int //что бы при отсылке команд ID уникальны

func NewManager(address string) *Manager {
	//address 10.10.10.10:5038
	client := Manager{Address: address}
	client.Events_chan = make(chan textproto.MIMEHeader, 50)
	client.Responce_chan = make(chan textproto.MIMEHeader, 50)
	return &client
}

func (client *Manager) Dial() (err error) {
	client.Conn, err = textproto.Dial("tcp", client.Address)
	return err
}

func (client *Manager) Stop() (err error) {
	err = client.Conn.Close()
	return err
}

func (client *Manager) SendAction(action string) (id int, err error) {
	actionID += 1 //каждый раз при запуске увеличиваем глобальную перем
	//fmt.Printf("ActionID: %d\r\n" + action, actionID)
	//Conn.Cmd сама добавляет \r\n
	_, err = client.Conn.Cmd("ActionID: %d\r\n"+action, actionID)
	return actionID, err
}

func (client *Manager) SendActionWaitResponse(action string) (resp textproto.MIMEHeader, err error) {
	a_id, err := client.SendAction(action)
	for {
		resp := <-client.Responce_chan
		if resp.Get("ActionID") == strconv.Itoa(a_id) {
			return resp, err
		}
	}
}

func (client *Manager) StartListen() {
	//	var s string
	for {
		//		if e, err := client.Conn.ReadLine(); err == nil {
		//			if e == "" {
		//				fmt.Println(s)
		//				client.Enents_string_chan <- s
		//				s = ""
		//			} else {
		//				s += e + "\t"
		//			}
		//		}
		if e, err := client.Conn.ReadMIMEHeader(); err == nil {
			//fmt.Println(e)
			if _, ok := e["Event"]; ok {
				//fmt.Println("event", e)
				client.Events_chan <- e
				if len(client.Events_chan)+2 > cap(client.Responce_chan) {
					<-client.Events_chan
				}
			} else if _, ok := e["Response"]; ok {
				//fmt.Println("Response", e)
				//что бы канал не переполнялся
				client.Responce_chan <- e
				if len(client.Responce_chan)+2 > cap(client.Responce_chan) {
					<-client.Responce_chan
				}
			} else {
				panic(e)
			}
		}
	}
}

func (client *Manager) Filter_Dial_event() {
	for {
		e := <-client.Events_chan
		if e.Get("Event") == "Dial" {
			fmt.Println(e.Get("Channel"))
		}
	}
}

func (client *Manager) Send_to_web(out_chan chan string) {
	for {
		e := <-client.Events_chan
		all_event := ""
		for k, v := range e {
			all_event += fmt.Sprintf("%s:%s<br>", k, v)
		}
		s := "<td><details><summary>" + e.Get("Event") + "<br>" + e.Get("Dialstring") + "</summary>" +
			all_event + "</details></td>"

		//		s := "<td>" + e.Get("Event") + "</td>" +
		//			"<td><details><summary>" + all_event + "</td>"

		//			"<td>" + e.Get("Uniqueid") + "</td>" +
		//			"<td>" + e.Get("Context") + "</td>" +
		//			"<td>" + e.Get("Exten") + "</td>" +
		//			"<td>" + e.Get("Extension") + "</td>" +
		//			"<td>" + e.Get("Priority") + "</td>" +
		//			"<td>" + e.Get("Application") + "</td>" +
		//			"<td>" + e.Get("Appdata") + "</td>" +
		//			"<td>" + e.Get("Channel") + "</td>" +
		//			"<td>" + e.Get("Calleridnum") + "</td>" +
		//			"<td>" + e.Get("Calleridname") + "</td>" +
		//			"<td>" + e.Get("Channelstate") + "</td>" +
		//			"<td>" + e.Get("Channelstatedesc") + "</td>" +
		//			"<td>" + e.Get("Variable") + "</td>" +
		//			"<td>" + e.Get("Value") + "</td>" +
		//			"<td>" + e.Get("Peer") + "</td>" +
		//			"<td>" + e.Get("Peerstatus") + "</td>" +
		//			"<td>" + e.Get("Dialstatus") + "</td>" +
		//			"<td>" + e.Get("Subevent") + "</td>"
		out_chan <- "<td>" + s + "</td>"
	}
}
