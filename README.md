# gomanyssh
## go运维平台,支持一万台主机同时并发执行脚本

使用方法

1.下载gomanyssh到任意目录

2.在gomanyssh同一目录创建一个ip.txt文件,格式有3种,例子以下
```
  192.168.1.1  22  root  123456   
  192.168.1.2  22  root  key
  192.168.1.2  22  root  key      abcdef


  ip   端口   用户名   密码   
  ip   端口   用户名   key(证书固定写成key)   证书无密码这个字段不用填
  ip   端口   用户名   key(证书固定写成key)   证书密码

  
  注意点:如果使用证书,请把证书的名字改为id_rsa,并放到和gomanyssh同一目录
```


3.执行命令,-c 后面写需要执行的命令即可 
```
  ./gomanyssh -c date
```

4.如果想先传输脚本到远程主机再执行,步骤以下
```
  ./gomanyssh -s 1.sh -d /tmp
```

5.再执行远程主机脚本 
```
  ./gomanyssh -c /tmp/1.sh
```



其它没有了,万台主机同时并发执行任务,可以用ulimit -n 102400把打开文件数设置大一点即可。

有任何意见,BUG,想法，欢迎大家一起来这里交流.


[运维架构](http://www.yunweijiagou.com/forum.php)

