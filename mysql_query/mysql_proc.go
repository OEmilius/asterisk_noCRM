package mysql_query

/*
mysql> CREATE USER aster_reader@'%' IDENTIFIED BY "web2secret";
mysql> grant select on asterisk.* TO aster_reader@"%";

CREATE USER web_front@'%' IDENTIFIED BY "web2pysecret";
GRANT ALL ON asterisk.* TO web_front@"%";

insert into `subscribers` set  number='151', enable='no', contract='офисный', litcevoy_schet='278346972165', metod_raschetov='Авансовый', blockirovka='Не блокирован', profit='2.3794', create_date_contract='26.03.2010 11:02:56', activation_date_contract='01.09.2014 11:29:17', tarif='Санкт-Петербург - БИЗНЕС СЕТЬ ОНЛАЙН_ПОМ';


./web2py.py --interfaces=172.31.22.153:8000

tester@asdf.as
te$termenow2
*/

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const PING_DURATION = 90 // интервал пингов MYSQL

//запустить и не закрывать коннект и пинговать периодически, что бы не отвалилось
func Get_sql_db() *sql.DB {
	fmt.Println("connect to mysql")
	db, err := sql.Open("mysql", "aster_reader:web2secret@tcp(52.17.173.106:3306)/asterisk")
	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}
	fmt.Println("connection to mysql successfull")
	//defer db.Close()
	go func() {
		ticker := time.NewTicker(PING_DURATION * time.Second)
		for {
			if _, ok := <-ticker.C; ok {
				err := db.Ping()
				if err != nil {
					fmt.Println("*****")
					panic(err.Error())
				}
				fmt.Println("ping MySQL db successfull")
				//web.WebPrint("ping ora db successfull")
			}
		}
	}()

	//	stmt, err := db.Prepare("select metod_raschetov,blockirovka from subscribers where number = ?")
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	return db
}

func Qery_sql(db *sql.DB, number string) (s string) {
	//	stmt, err := db.Prepare("select metod_raschetov,blockirovka from subscribers where number = ?")
	//	if err != nil {
	//		fmt.Println("error in stmt", err)
	//	}
	q := fmt.Sprintf("select dtmf, duration, pdd, count, tarif, contract, litcevoy_schet, activation_date_contract, metod_raschetov, blockirovka, profit from subscribers where number = %s", number)
	//q := fmt.Sprintf("select * from subscribers where number = %s", number)
	//fmt.Println(q)
	rows, err := db.Query(q)
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		var dtmf, duration, pdd, count, tarif, contract, litcevoy_schet, activation_date_contract, metod_raschetov, blockirovka, profit string
		err = rows.Scan(&dtmf, &duration, &pdd, &count, &tarif, &contract, &litcevoy_schet, &activation_date_contract, &metod_raschetov, &blockirovka, &profit)
		//fmt.Println(metod_raschetov, blockirovka)
		if err != nil {
			fmt.Println("++++++++error scan rows", err)
			panic(err)
		}
		fmt.Println("=======SQL find ", litcevoy_schet)
		//s = fmt.Sprintf("%s %s %s %s %s", calldate, number, metod_raschetov, tarif, blockirovka, success, count)
		s += "<table bordercolor=#008800><tr>"
		s += "<td>номер</td>"
		s += "<td>dtmf</td>"
		s += "<td>длительность</td>"
		s += "<td>pdd</td>"
		s += "<td>кол-во</td>"
		s += "<td>тариф</td>"
		s += "<td>контракт</td>"
		s += "<td>лицевой</td>"
		s += "<td>активация</td>"
		s += "<td>метод расчетов</td>"
		s += "<td>блокировка</td>"
		s += "<td>начислено</td>"
		s += "</tr>"
		s += "<tr>"
		s += fmt.Sprintf("<td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td>", number, dtmf, duration, pdd, count, tarif, contract, litcevoy_schet, activation_date_contract, metod_raschetov, blockirovka, profit)
		s += "</tr>"
		s += "</table>"
		//s += fmt.Sprintf("%s %s %s %s %s", calldate, number, metod_raschetov, tarif, blockirovka, success, count)
	}
	return s
}

//func main() {
//	db := Get_sql_db()
//	//fmt.Println(db)
//	r := Qery_sql(db, "79811545299")
//	fmt.Println(r)
//	r = Qery_sql(db, "7981154529")
//	fmt.Println("dfd", r)
//	fmt.Scanln()
//}
