# asterisk_noCRM
connect to asterisk using AMI, push new incomming calls information to web page using websocket, allow to call from web page using originate command.

Configuration in config.json
{

	"Mysql_connect_string": "root@tcp(localhost:3306)/freeswitch",
	"WebHost_port": "127.0.0.1:8888",   // ip address and port to start web server
	"Ami_host_port": "192.168.0.60",  // asterisk AMI socket 
	"Ami_Username": "ami_user",    
	"Ami_Secret": "password",
	"Context": "office", // context to create outgoing call
	"end_config": "end"
	
}

