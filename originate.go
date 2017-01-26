/*TODO: 2 инпута и префикс 09
кнопочку redial в истории звонков
кнопку hangup
кнопку reconnect to ami

*/
package main

import (
	"asterisk_noCRM2/ami"
	"asterisk_noCRM2/web_print_json"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Call struct {
	Intype       string
	Calleridnum  string
	Dialstring   string
	Calleridname string
	Uniqueid     string
	DateTime     string
	Status       string
}

var Clients = make(map[string]*Cleint)

type Cleint struct {
	Phone  string
	Web_ch chan string
}

func ReadConfig(fname string) Config {
	file, err := os.Open(fname)
	if err != nil {
		dbg.Println("opening config file error", err)
		panic(err)
	}
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		dbg.Println("error reading config file", err)
		panic(err)
	}
	//fmt.Println("config=", config)
	return config
}

var cfg Config

type Config struct {
	Mysql_connect_string string
	WebHost_port         string
	Ami_host_port        string
	Ami_Username         string
	Ami_Secret           string
	Context              string
}

func main() {
	dbg.Println("start")
	cfg = ReadConfig("config.json")
	dbg.Printf("#%v\r\n", cfg)
	//web_print_json.WebHost_port = "192.168.20.150:8088"
	web_print_json.WebHost_port = cfg.WebHost_port
	m := ami.NewManager(cfg.Ami_host_port)
	if err := m.Dial(); err == nil {
		go m.StartListen()
	}
	_, err := m.SendAction("Action: Login\r\nUsername: " + cfg.Ami_Username + "\r\nSecret: " + cfg.Ami_Secret + "\r\n")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	go web_print_json.Start()
	go Read_command(web_print_json.Cmd_ch, m)
	dial_events_chan := make(chan Call)
	go Dial_begin_finder(m, dial_events_chan)
	User_events_finder(dial_events_chan, "150")
	//fmt.Scanln()
}

func Dial_begin_finder(m *ami.Manager, out_ch chan Call) {
	for {
		e := <-m.Events_chan
		if e.Get("Event") == "DialBegin" {
			fmt.Println(e.Get("Calleridname"), e.Get("Calleridnum"), "call to:", e.Get("Dialstring"))
			dbg.Println(e.Get("Uniqueid"))
			c := Call{Calleridnum: e.Get("Calleridnum"),
				Dialstring:   e.Get("Dialstring"),
				Calleridname: e.Get("Calleridname"),
				Uniqueid:     e.Get("Uniqueid"),
				Status:       "DialBegin"}
			out_ch <- c
		} else if e.Get("Event") == "DialEnd" {
			dbg.Println("...end", e.Get("Calleridname"), e.Get("Calleridnum"), "end call with reason", e.Get("Dialstatus"))
			c := Call{Calleridnum: e.Get("Calleridnum"),
				Dialstring:   e.Get("Dialstring"),
				Calleridname: e.Get("Calleridname"),
				Uniqueid:     e.Get("Uniqueid"),
				Status:       "DialEnd"}
			dbg.Println("...", c)
			out_ch <- c
		}
	}
}

//TODO сделать таблицу, и аплоад в python

func User_events_finder(in_chan chan Call, userid string) {
	for {
		call := <-in_chan
		//s := "<tr><td>" + call.Calleridnum + "</td></tr>"
		//web.Data_chan <- s
		call.Intype = "Call"
		phone := call.Dialstring
		if client, ok := Clients[phone]; ok {
			call.DateTime = time.Now().String()
			jmsg, err := json.Marshal(call)
			if err != nil {
				panic(err)
			}
			dbg.Println("***", string(jmsg))
			dbg.Println("=====len", len(client.Web_ch))
			dbg.Println("=====cap", cap(client.Web_ch))
			if cap(client.Web_ch) != 0 {
				client.Web_ch <- string(jmsg)
			}
		}
	}
}

var ID_Phone = make(map[int]string)

func Read_command(in_ch chan web_print_json.Cmd, amiManager *ami.Manager) {
	dbg.Println("start read commands")
	for {
		cmd := <-in_ch
		dbg.Printf("%#v\r\n", cmd)
		wc := web_print_json.Clients[cmd.Id]
		id := cmd.Id
		if strings.Index(cmd.Msg, ":") > 0 {
			command := cmd.Msg[:strings.Index(cmd.Msg, ":")]
			cmdarg := cmd.Msg[strings.Index(cmd.Msg, ":")+1:]
			dbg.Println("client=", cmd.Id, "cmd=", cmd.Msg, "arg=", cmdarg)
			switch command {
			case "newWebClient":
				dbg.Println("case New web Client id=", cmd.Id)
			case "myphone":
				dbg.Println("**case my phone=", cmdarg)
				myphone := cmdarg
				client := Cleint{}
				client.Web_ch = web_print_json.Clients[id].Ans_ch
				Clients[myphone] = &client
				Clients[myphone].Phone = cmdarg
				dbg.Println(Clients[myphone])
				ID_Phone[id] = myphone
			case "dial":
				from := cmdarg[:strings.Index(cmdarg, "t")]
				to := cmdarg[strings.Index(cmdarg, "t")+1:]
				dbg.Println(from, "|", to)
				amiManager.SendAction(PrepOriginate(from, to))
			case "hangup":
				phone := cmdarg
				h := fmt.Sprintf("Action: Hangup\r\nChannel:/^SIP/%s-.*$/\r\n", phone)
				id, err := amiManager.SendAction(h)
				dbg.Println(id, err)
			case "close_ws":
				dbg.Println("******close ws")
				p := ID_Phone[id]
				delete(Clients, p)

				_ = wc
				_ = id
			}
		}
	}
}

func PrepOriginate(calling, called string) string {
	return fmt.Sprintf("Action: Originate\r\n"+
		"Channel:SIP/%s\r\n"+
		"Context:%s\r\n"+
		"Exten:%s\r\n"+
		"Priority:1\r\n"+
		//"Callerid:150\r\n"+
		"Variable: SIPADDHEADER=\"Call-Info:;answer-after=0\""+
		"Timeout:30000\r\n", calling, cfg.Context, called)
}

var dbg debugging = true

type debugging bool

func (dbg debugging) Printf(format string, args ...interface{}) {
	if dbg {
		fmt.Printf(format, args...)
	}
}

func (dbg debugging) Println(args ...interface{}) {
	if dbg {
		fmt.Println(args...)
	}
}
