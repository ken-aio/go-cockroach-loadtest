package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	dbUser = flag.String("U", "root", "db user")
	dbHost = flag.String("h", "lcoalhost", "db host name or ip addr")
	// dbPassword = flag.String("p", "", "db password")
	dbPort      = flag.String("P", "26257", "db port")
	dbName      = flag.String("d", "test", "db name")
	reqNum      = flag.Int("n", 1000, "total request num")
	parallelNum = flag.Int("t", 20, "parallel number")
	debug       = flag.Bool("debug", false, "true if debug mode")
)

// Test db struct
type Test struct {
	ID        int
	Code      string
	Text      string
	IsTest    bool
	CreatedAt time.Time
}

func main() {
	flag.Parse()
	var sema chan int = make(chan int, *parallelNum)
	begin := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < *reqNum; i++ {
		wg.Add(1)
		go load(sema, &wg)
	}
	wg.Wait()
	end := time.Now()
	fmt.Println(end.Sub(begin))
	count := selectCount()
	fmt.Printf("insert num = %+v, select count = %v\n", *reqNum, count)
}

func load(sema chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	sema <- 1
	defer func() { <-sema }()
	insert()
}

func connect() *sql.DB {
	//dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", *dbUser, *dbPassword, *dbName, *dbHost, *dbPort)
	dbinfo := fmt.Sprintf("user=%s dbname=%s host=%s port=%s sslmode=disable", *dbUser, *dbName, *dbHost, *dbPort)
	if *debug {
		log.Printf("dbinfo = %+v\n", dbinfo)
	}
	db, err := sql.Open("postgres", dbinfo)
	fatalIfErr(err)
	return db
}

func insert() int64 {
	sql := "insert into test(code, text, is_test, created_at) values($1, $2, $3, $4)"
	db := connect()
	defer db.Close()
	code := generateUID()
	result, err := db.Exec(sql, code, "test", true, time.Now())
	fatalIfErr(err)
	rows := int64(-1)
	if *debug {
		rows, err := result.RowsAffected()
		fatalIfErr(err)
		lastID, err := result.LastInsertId()
		// fatalIfErr(err) => cockroachdbの場合は取れない
		fmt.Printf("rows = %+v, last_insert_id = %+v\n", rows, lastID)
	}
	return rows
}

func selectList() []*Test {
	sql := "select * from test"
	db := connect()
	defer db.Close()

	rows, err := db.Query(sql)
	fatalIfErr(err)
	defer rows.Close()

	tests := make([]*Test, 0)
	for rows.Next() {
		t := &Test{}
		if err := rows.Scan(&t.Code, &t.Text, &t.IsTest, &t.CreatedAt); err != nil {
			fatalIfErr(err)
		}
		tests = append(tests, t)
	}
	return tests
}

func selectCount() int {
	sql := "select count(*) from test"
	db := connect()
	defer db.Close()

	rows, err := db.Query(sql)
	fatalIfErr(err)
	defer rows.Close()

	var count int
	rows.Next()
	if err := rows.Scan(&count); err != nil {
		fatalIfErr(err)
	}
	return count
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// https://qiita.com/shinofara/items/5353df4f4fbdaae3d959
func generateUID() string {
	buf := make([]byte, 10)

	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	str := fmt.Sprintf("%d%x", time.Now().Unix(), buf[0:10])
	return hex.EncodeToString([]byte(str))
}
