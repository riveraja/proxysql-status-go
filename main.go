package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/cheynewallace/tabby"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var queryString, tableType string

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func pcheck(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	//os.Remove("./test.txt")

	//writeFile, err := os.OpenFile("./test.txt", os.O_WRONLY | os.O_CREATE, 0666)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer writeFile.Close()
	var userName, passwd string
	var sPort int

	// Declare flags
	boolStats := flag.Bool("stats", false, "Generate stats data")
	boolGR := flag.Bool("groupreplication", false, "Show Group Replication HostGroups")
	boolFiles := flag.Bool("files", false, "Show file contents")
	boolAll := flag.Bool("all", false, "Show all")
	boolRuntime := flag.Bool("runtime", false, "Show runtime tables only")
	flag.StringVar(&userName, "user", "admin", "ProxySQL username")
	flag.StringVar(&passwd, "password", "admin", "ProxySQL password")
	flag.IntVar(&sPort, "port", 6032, "ProxySQL port")

	flag.Parse()
	// End declare flags

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:%d)/main", userName, passwd, sPort)

	strVal := strconv.FormatBool(*boolRuntime)

	var err error

	db, err = sql.Open("mysql", dsn)
	check(err)
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	hostName, err := os.Hostname()
	check(err)

	mysqlserverCount, err := db.Query("select count(*) from mysql_servers")
	check(err)
	defer mysqlserverCount.Close()

	mysqluserCount, err := db.Query("select count(*) from mysql_users")
	check(err)
	defer mysqluserCount.Close()

	runtimemysqluserCount, err := db.Query("select count(*) from runtime_mysql_users")
	check(err)
	defer runtimemysqluserCount.Close()

	runtimemysqlserverCount, err := db.Query("select count(*) from runtime_mysql_servers")
	check(err)
	defer runtimemysqlserverCount.Close()

	timeNow := time.Now()
	fmt.Println("########## ProxySQL Summary Report ##########")
	fmt.Printf("Date/Time:       %s\n", timeNow.Format(time.RFC1123))
	fmt.Printf("Hostname:        %s\n", hostName)

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

	fmt.Printf("MySQL Users:     %d / %d\n", imysqluserCount, iruntimemysqluserCount)

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

	fmt.Println("\n########## ProxySQL MySQL Servers ##########")

	tableType = funcTabletype(strVal)

	queryString = fmt.Sprintf("select hostgroup_id,hostname,port,status,weight,compression,max_connections,max_replication_lag,use_ssl,max_latency_ms,comment from %smysql_servers order by hostgroup_id", tableType)
	srows, err := db.Query(queryString)
	check(err)
	defer srows.Close()

	var hid, port, wt, comp, maxcon, maxrepl, usessl, maxlat int
	var hname, st, comment string
	t := tabby.New()
	t.AddHeader("HG", "Hostname", "Port", "Status", "Weight", "Compression", "Max Conn", "Max Repl Lag", "Use SSL", "Max Latency", "Comment")
	for srows.Next() {
		if err := srows.Scan(&hid, &hname, &port, &st, &wt, &comp, &maxcon, &maxrepl, &usessl, &maxlat, &comment); err != nil {
			panic(err)
		}
		t.AddLine(hid, hname, port, st, wt, comp, maxcon, maxrepl, usessl, maxlat, comment)
	}

	t.Print()

	fmt.Println("\n########## ProxySQL MySQL Users ##########")

	tableType = funcTabletype(strVal)

	queryString = fmt.Sprintf("select username,active,use_ssl,default_hostgroup,default_schema,schema_locked,transaction_persistent,fast_forward,backend,frontend,max_connections from %smysql_users", tableType)
	irows, err := db.Query(queryString)
	pcheck(err)
	defer irows.Close()

	var uname, defHG string
	var defSchema sql.NullString
	var ndefSchema string
	var active, useSSL, schemaLocked, trxPersistent, fastFWD, backend, frontend, maxconn int
	m := tabby.New()
	m.AddHeader("Username", "Active", "Use SSL", "Default HG", "Default Schema", "Schema Locked", "Trx Persistent", "Fast Fwd", "Backend", "Frontend", "Max Conn")
	for irows.Next() {
		if err := irows.Scan(&uname, &active, &useSSL, &defHG, &defSchema, &schemaLocked, &trxPersistent, &fastFWD, &backend, &frontend, &maxconn); err != nil {
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

	tableType = funcTabletype(strVal)

	queryString = fmt.Sprintf("select id, active, interval_ms, filename, arg1, arg2, arg3, arg4, arg5, comment from %sscheduler", tableType)
	sched, err := db.Query(queryString)
	pcheck(err)
	defer sched.Close()

	var id, intervalMS int
	var filename, narg1, narg2, narg3, narg4, narg5 string
	var arg1, arg2, arg3, arg4, arg5 sql.NullString
	for sched.Next() {
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
		fmt.Printf("\nComment:        %s\n", comment)
	}

	fmt.Println("\n########## MySQL Replication Hostgroups ##########")

	tableType = funcTabletype(strVal)

	queryString = fmt.Sprintf("select * from %smysql_replication_hostgroups", tableType)
	rhg, err := db.Query(queryString)
	pcheck(err)
	defer rhg.Close()

	var writehg, readhg int
	s := tabby.New()
	s.AddHeader("Writer HG", "Reader HG", "Comment")
	for rhg.Next() {
		if err := rhg.Scan(&writehg, &readhg, &comment); err != nil {
			panic(err)
		}
		s.AddLine(writehg, readhg, comment)
	}

	s.Print()

	fmt.Println("\n########## MySQL Query Rules ##########")

	tableType = funcTabletype(strVal)

	queryString = fmt.Sprintf("select rule_id,active,username,schemaname,digest,match_digest,match_pattern,negate_match_pattern,replace_pattern,destination_hostgroup,apply,comment from %smysql_query_rules", tableType)
	qr, err := db.Query(queryString)
	pcheck(err)
	defer qr.Close()

	var ruleID, actve, nmatchPattern, destHg, mapply int
	var schemaName, digest, matchDigest, matchPattern, replacePattern, mcomment string
	var schemaNamex, digestx, matchDigestx, matchPatternx, replacePatternx, mcommentx sql.NullString
	q := tabby.New()
	q.AddHeader("RuleID", "Active", "User", "Schema", "Digest", "MatchDigest", "MatchPattern", "NegatePattern", "ReplacePattern", "DestHG", "Apply", "Comment")
	for qr.Next() {
		if err := qr.Scan(&ruleID, &actve, &userName, &schemaNamex, &digestx, &matchDigestx, &matchPatternx, &nmatchPattern, &replacePatternx, &destHg, &mapply, &mcommentx); err != nil {
			panic(err)
		}
		if schemaNamex.Valid {
			schemaName = schemaNamex.String
		} else {
			schemaName = "NULL"
		}
		if digestx.Valid {
			digest = digestx.String
		} else {
			digest = "NULL"
		}
		if matchDigestx.Valid {
			matchDigest = matchDigestx.String
		} else {
			matchDigest = "NULL"
		}
		if matchPatternx.Valid {
			matchPattern = matchDigestx.String
		} else {
			matchPattern = "NULL"
		}
		if replacePatternx.Valid {
			replacePattern = replacePatternx.String
		} else {
			replacePattern = "NULL"
		}
		if mcommentx.Valid {
			mcomment = mcommentx.String
		} else {
			mcomment = "NULL"
		}
		q.AddLine(ruleID, actve, userName, schemaName, digest, matchDigest, matchPattern, nmatchPattern, replacePattern, destHg, mapply, mcomment)
	}
	q.Print()

	fmt.Println("\n########## ProxySQL Global Variables ##########")

	tableType = funcTabletype(strVal)

	queryString = fmt.Sprintf("select * from %sglobal_variables", tableType)
	rows, err := db.Query(queryString)
	check(err)
	defer rows.Close()

	var name, val string
	for rows.Next() {
		if err := rows.Scan(&name, &val); err != nil {
			panic(err)
		}
		fmt.Printf("%s: %s\n", name, val)
	}

	if *boolGR == true || *boolAll == true {
		showGR(strVal)
	}

	if *boolStats == true || *boolAll == true {
		showStats()
	}

	if *boolFiles == true || *boolAll == true {
		showFiles()
	}

	fmt.Println("\n#### End ####")
	//#### Cleanup Section ####
	//os.Remove("./statusfile.txt")
}

func showStats() {

	fmt.Println("\n########## ProxySQL Stats MySQL Connection Pool ##########")

	sscpl, err := db.Query("SELECT hostgroup, srv_host, status, ConnUsed, ConnFree, ConnOK, ConnERR FROM stats.stats_mysql_connection_pool WHERE ConnUsed+ConnFree > 0 ORDER BY hostgroup, srv_host")
	check(err)
	defer sscpl.Close()

	var hostgroup, connUsed, connFree, connOK, connERR int
	var srvHost, status string
	cpl := tabby.New()
	cpl.AddHeader("HG", "Srv Host", "Status", "ConnUsed", "ConnFree", "ConnOK", "ConnERR")
	for sscpl.Next() {
		if err := sscpl.Scan(&hostgroup, &srvHost, &status, &connUsed, &connFree, &connOK, &connERR); err != nil {
			panic(err)
		}
		cpl.AddLine(hostgroup, srvHost, status, connUsed, connFree, connOK, connERR)
	}
	cpl.Print()

}

func showGR(r string) {

	fmt.Println("\n########## MySQL Group Replication Hostgroups ##########")

	tableType = funcTabletype(r)

	queryString = fmt.Sprintf("select * from %smysql_group_replication_hostgroups", tableType)
	grhg, err := db.Query(queryString)
	pcheck(err)
	defer grhg.Close()

	var writehg, bkwritehg, readerhg, offlinehg, active, maxwriters, wrrd, maxtrx int
	var comment string
	g := tabby.New()
	g.AddHeader("Writer HG", "Backup Writer HG", "Reader HG", "Offline HG", "Active", "Max Writers", "Writer is reader", "Max Trx Behind", "Comment")
	for grhg.Next() {
		if err := grhg.Scan(&writehg, &bkwritehg, &readerhg, &offlinehg, &active, &maxwriters, &wrrd, &maxtrx, &comment); err != nil {
			panic(err)
		}
		g.AddLine(writehg, bkwritehg, readerhg, offlinehg, active, maxwriters, wrrd, maxtrx, comment)
	}
	g.Print()

}

func showFiles() {

	fmt.Println("\n########## ProxySQL Files ##########")

	fmt.Println("\nFile: /etc/proxysql-admin.cnf")
	m, err := ioutil.ReadFile("/etc/proxysql-admin.cnf")
	if err != nil {
		fmt.Printf("Failed to %s\n", err)
	}
	fmt.Print(string(m))

	fmt.Println("\nFile: /var/lib/proxysql/host_priority.conf")
	s, err := ioutil.ReadFile("/var/lib/proxysql/host_priority.conf")
	if err != nil {
		fmt.Printf("Failed to %s\n", err)
	}
	fmt.Print(string(s))
}

func funcTabletype(r string) string {
	if r == "true" {
		return "runtime_"
	}
	return ""
}
