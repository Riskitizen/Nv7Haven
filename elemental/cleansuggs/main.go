package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql
)

const (
	dbUser     = "u51_iYXt7TBZ0e"
	dbPassword = "W!QnD2u896yo.J4fww9X.h+J"
	dbName     = "s51_nv7haven"
)

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

var sg sync.WaitGroup

func main() {
	db, err := sql.Open("mysql", dbUser+":"+dbPassword+"@tcp(c.filipk.in:3306)/"+dbName)
	handle(err)
	defer db.Close()

	fmt.Println("Connected")

	res, err := db.Query("SELECT elem1, elem2 FROM elem_combos WHERE 1")
	handle(err)
	defer res.Close()
	var belem1 string
	var belem2 string
	var name string

	for res.Next() {
		err = res.Scan(&belem1, &belem2)
		handle(err)
		go func() {
			elem1 := belem1
			elem2 := belem2
			rs, err := db.Query("SELECT elem3 FROM sugg_combos WHERE (elem1=? AND elem2=?) OR (elem1=? AND elem2=?)", elem1, elem2, elem2, elem1)
			handle(err)
			for rs.Next() {
				err = rs.Scan(&name)
				handle(err)
				_, err = db.Exec("DELETE FROM suggestions WHERE name=?", name)
				handle(err)
				fmt.Println(name)
			}
			rs.Close()

			_, err = db.Exec("DELETE FROM sugg_combos WHERE (elem1=? AND elem2=?) OR (elem1=? AND elem2=?)", elem1, elem2, elem2, elem1)
			handle(err)
			sg.Done()
		}()
		sg.Add(1)
		time.Sleep(time.Second / 10)
	}
	sg.Wait()
}
