package main

import (
	"fmt"
	"net/http"
	"log"
	"os"
	"io/ioutil"
	"bufio"
	"strings"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

var Dow = []string{
"AAPL",	"AXP", "BA", 	"CAT", 	"CSCO",
"CVX", 	"KO", 	"DD", 	"XOM", 	"GE",
"GS", 	"HD", 	"IBM", 	"INTC", "JNJ",
"JPM", 	"MCD", 	"MMM", 	"MRK", 	"MSFT",
"NKE", 	"PFE", 	"PG", 	"TRV", 	"UNH",
"UTX", 	"V", 	"VZ", 	"WMT", 	"DIS",
}

func History (ticker string, p string) string {
	fs := "http://ichart.finance.yahoo.com/table.csv?s=%s&c=1970"
	u := fmt.Sprintf(fs, ticker)

	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	file, err := os.Create(p)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	file.Write(c)
	return p
}

func HistoryToSQL (ticker string)  {
	fs := "http://ichart.finance.yahoo.com/table.csv?s=%s&c=1970"
	u := fmt.Sprintf(fs, ticker)
	//fmt.Println(u)
	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	r := bufio.NewScanner(resp.Body)
	q := "INSERT INTO history VALUES ('" + ticker + "', DATE '%s',%v,%v,%v,%v,%v,%v);"

	db, err := sql.Open("mysql", "root:password@/stocks")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r.Scan()
	for r.Scan() {
		i := strings.Split(r.Text(), ",")
		_, err := db.Exec(fmt.Sprintf(q, i[0],i[1],i[2],i[3],i[4],i[5],i[6]))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func IsTradingTime()  bool{
	hr := time.Now().Hour()
	min := time.Now().Minute()
	day := time.Now().Weekday()

	return (hr > 9 && min > 30) && (hr < 5 && min < 30) && !(day == time.Sunday || day == time.Saturday)
}

func NowToSQL(ticker string)  {
	fs := "http://finance.yahoo.com/d/quotes.csv?s=%s&f=t1l1"
	u := fmt.Sprintf(fs, ticker)

	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	r := bufio.NewScanner(resp.Body)

	q := "INSERT INTO daily VALUES ('" + ticker + "', TIMESTAMP '%s', %v);"

	db, err := sql.Open("mysql", "root:password@/stocks")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	r.Scan()
	i := strings.Split(r.Text(), ",")

	t, err := time.Parse(`2006-01-02 "3:04pm" MST`,
		strings.Split(time.Now().String(), " ")[0] + " " + i[0] + " EST")
	if err != nil {
		log.Fatal(err)
	}


	_, err = db.Exec(
		fmt.Sprintf(q, t.Format(
			"2006-01-02 15:04:05"), i[1]))
	if err != nil {
		log.Fatal(err)
	}

}

func GetDowComp()  {
	dow := Dow

	for i := 27; i < len(dow); i++{
		fmt.Println("getting " + dow[i])
		HistoryToSQL(dow[i])
	}
}

func main() {

	for i := 0; i < len(Dow); i++ {
		for true {
			go NowToSQL(Dow[i])
			time.Sleep(1800 * time.Millisecond)
		}
	}
	NowToSQL("GOOG")
}
