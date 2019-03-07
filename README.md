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
