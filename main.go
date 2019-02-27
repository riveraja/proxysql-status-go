package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cheynewallace/tabby"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	//os.Remove("./test.txt")

	//writeFile, err := os.OpenFile("./test.txt", os.O_WRONLY | os.O_CREATE, 0666)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer writeFile.Close()

	db, err := sql.Open("mysql", "admin:admin@tcp(127.0.0.1:6032)/main")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hostName, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	mysqlserverCount, err := db.Query("select count(*) from mysql_servers")
	if err != nil {
		log.Fatal(err)
	}
	defer mysqlserverCount.Close()

	mysqluserCount, err := db.Query("select count(*) from mysql_users")
	if err != nil {
		log.Fatal(err)
	}
	defer mysqluserCount.Close()

	runtimemysqluserCount, err := db.Query("select count(*) from runtime_mysql_users")
	if err != nil {
		log.Fatal(err)
	}
	defer runtimemysqluserCount.Close()

	runtimemysqlserverCount, err := db.Query("select count(*) from runtime_mysql_servers")
	if err != nil {
		log.Fatal(err)
	}
	defer runtimemysqlserverCount.Close()

	timeNow := time.Now()
	fmt.Println("########## ProxySQL Summary Report ##########")
	fmt.Printf("Date/Time:             %s\n", timeNow.Format(time.RFC1123))
	fmt.Printf("Hostname:              %s\n", hostName)

	var imysqluserCount, iruntimemysqluserCount, imysqlserverCount, iruntimemysqlserverCount int

	for mysqluserCount.Next() {
		if err := mysqluserCount.Scan(&imysqluserCount); err != nil {
			panic(err)
		}
	}
	for runtimemysqluserCount.Next() {
		if err := runtimemysqluserCount.Scan(&iruntimemysqluserCount); err != nil {
			panic(err)
		}
	}

	fmt.Printf("MySQL Users:    %d / %d\n", imysqluserCount, iruntimemysqluserCount)

	for mysqlserverCount.Next() {
		if err := mysqlserverCount.Scan(&imysqlserverCount); err != nil {
			panic(err)
		}
	}
	for runtimemysqlserverCount.Next() {
		if err := runtimemysqlserverCount.Scan(&iruntimemysqlserverCount); err != nil {
			panic(err)
		}
	}

	fmt.Printf("MySQL Servers:   %d / %d\n", imysqlserverCount, iruntimemysqlserverCount)

	fmt.Println("\n########## ProxySQL Global Variables ##########")

	rows, err := db.Query("select * from runtime_global_variables where variable_name like 'mysql-%'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, val string
		if err := rows.Scan(&name, &val); err != nil {
			panic(err)
		}
		fmt.Printf("%s: %s\n", name, val)
	}

	fmt.Println("\n########## ProxySQL MySQL Servers ##########")

	srows, err := db.Query("select hostgroup_id,hostname,port,weight,compression,max_connections,max_replication_lag,use_ssl,max_latency_ms,comment from mysql_servers order by hostgroup_id")
	if err != nil {
		log.Fatal(err)
	}
	defer srows.Close()

	t := tabby.New()
	t.AddHeader("HG", "Hostname", "Port", "Weight", "Compression", "Max Conn", "Max Repl Lag", "Use SSL", "Max Latency", "Comment")
	for srows.Next() {
		var hid, port, wt, comp, maxcon, maxrepl, usessl, maxlat int
		var hname, comment string
		if err := srows.Scan(&hid, &hname, &port, &wt, &comp, &maxcon, &maxrepl, &usessl, &maxlat, &comment); err != nil {
			panic(err)
		}
		t.AddLine(hid, hname, port, wt, comp, maxcon, maxrepl, usessl, maxlat, comment)
	}

	t.Print()

	fmt.Println("\n########## ProxySQL MySQL Users ##########")

	irows, err := db.Query("select username,active useSSL,default_hostgroup,default_schema,schema_locked,transaction_persistent,fast_forward,backend,frontend,max_connections from mysql_users")
	if err != nil {
		panic(err)
	}
	defer irows.Close()

	m := tabby.New()
	m.AddHeader("Username", "Active", "Use SSL", "Default HG", "Default Schema", "Schema Locked", "Trx Persistent", "Fast Fwd", "Backend", "Frontend", "Max Conn")
	for irows.Next() {
		var uname, defHG string
		var defSchema sql.NullString
		var ndefSchema string
		var active, useSSL, schemaLocked, trxPersistent, fastFWD, backend, frontend, maxconn int
		if err := irows.Scan(&uname, &active, useSSL, &defHG, &defSchema, &schemaLocked, &trxPersistent, &fastFWD, &backend, &frontend, &maxconn); err != nil {
			panic(err)
		}
		if defSchema.Valid {
			ndefSchema = defSchema.String
		} else {
			ndefSchema = "NULL"
		}
		m.AddLine(uname, active, useSSL, defHG, ndefSchema, schemaLocked, trxPersistent, fastFWD, backend, frontend, maxconn)
	}

	m.Print()

	fmt.Println("\n########## ProxySQL Scheduler ##########")

	sched, err := db.Query("select id, active, intervalMS, filename, arg1, arg2, arg3, arg4, arg5, comment from scheduler")
	if err != nil {
		panic(err)
	}
	defer sched.Close()

	for sched.Next() {
		var id, active, intervalMS int
		var filename, comment, narg1, narg2, narg3, narg4, narg5 string
		var arg1, arg2, arg3, arg4, arg5 sql.NullString
		if err := sched.Scan(&id, &active, &intervalMS, &filename, &arg1, &arg2, &arg3, &arg4, &arg5, &comment); err != nil {
			panic(err)
		}
		if arg1.Valid {
			narg1 = arg1.String
		} else {
			narg1 = "NULL"
		}
		if arg2.Valid {
			narg2 = arg2.String
		} else {
			narg2 = "NULL"
		}
		if arg3.Valid {
			narg3 = arg3.String
		} else {
			narg3 = "NULL"
		}
		if arg4.Valid {
			narg4 = arg4.String
		} else {
			narg4 = "NULL"
		}
		if arg5.Valid {
			narg5 = arg5.String
		} else {
			narg5 = "NULL"
		}
		fmt.Printf("\nScheduler ID:   %d", id)
		fmt.Printf("\nIs Active:      %d", active)
		fmt.Printf("\nInterval (ms):  %d", intervalMS)
		fmt.Printf("\nFilename:       %s", filename)
		fmt.Printf("\nArg1:           %s", narg1)
		fmt.Printf("\nArg2:           %s", narg2)
		fmt.Printf("\nArg3:           %s", narg3)
		fmt.Printf("\nArg4:           %s", narg4)
		fmt.Printf("\nArg5:           %s", narg5)
		fmt.Printf("\nComment:        %s", comment)
	}
	fmt.Println("\n#### End ####")
	//#### Cleanup Section ####
	//os.Remove("./statusfile.txt")
}