# proxysql-status-go
A proxysql-status alternative written in Golang

```
[root@proxysql1 proxysql-status-go]# time ./proxysql-status-go > status.txt

real	0m0.031s
user	0m0.006s
sys	0m0.012s
[root@proxysql1 proxysql-status-go]# time proxysql-status admin admin 127.0.0.1 6032 > status.txt

real	0m0.498s
user	0m0.265s
sys	0m0.202s
```

```
[root@proxysql1 proxysql-status-go]# ./proxysql-status-go --help
Usage of ./proxysql-status-go:
  -all
    	Show all
  -files
    	Show file contents
  -groupreplication
    	Show Group Replication HostGroups
  -password string
    	ProxySQL password (default "admin")
  -port int
    	ProxySQL port (default 6032)
  -runtime
    	Show runtime tables only
  -stats
    	Generate stats data
  -user string
    	ProxySQL username (default "admin")
```