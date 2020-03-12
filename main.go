package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gomanyssh/utils"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ipResultChanel = make(chan string,1000)

//存放ip
var file_map_ips = make(map[int]string)
//存放ip对应的端口号
var file_map_port = make(map[int]string)
//存放ip对应的用户名
var file_map_user = make(map[int]string)
//存放ip对应的密码
var file_map_password = make(map[int]string)
//存放ip对应的证书密码
var file_map_keyfilepassword = make(map[int]string)

//统计执行成功的ip数
var runSuccessNumber = 0

//脚本执行成功日志文件
var logFile *os.File
//脚本执行错误日志文件
var errLogFile *os.File
//当有还在运行的ip信息文件
var ipStillRunningFile *os.File


//通道返回的结果字符
var resultInfo string

//执行命令
var (
	sshCmd *string
	sshScpSourceFile *string
	sshScpDestPath *string
)


const (
	color_red = uint8(iota + 91)
	color_green		//	绿
	color_yellow		//	黄
	color_blue			// 	蓝
	color_magenta 		//	洋红
)


func main() {


	//开始计算运行时间
	startTime := time.Now()

	//处理用户输入的参数
	sshCmd = flag.String("c", "", "需要执行的命令,可以用;或&&连接一长串命令,暂不支持搭配-s或-d同时执行")
	sshScpSourceFile = flag.String("s", "", "需要复制到远程主机的文件,单独使用没有效果,必须搭配-d 同时使用")
	sshScpDestPath = flag.String("d", "", "需要复制到远程主机的路径,单独使用没有效果,必须搭配-s 同时使用")

	flag.Parse()

	if flag.NFlag() <= 0 {
		log.Fatal("请使用-c 设置远程执行命令,或使用-s -d 传输文件到远程主机")
	} else if flag.NFlag() == 1 {
		//如果只设置一个参数,要判断是不是-s，或-d
		if *sshScpSourceFile != "" || *sshScpDestPath != "" {
			log.Fatal("请先传可执行文件到远程主机,之后再执行,为了并发安全,暂不支持传文件和执行命令同时运行")
		}

	}
	//在实际生产环境中,先传文件,再执行,缓冲一下并发前的准备工作和准备
	//如果合在一起,并发出问题很麻烦
	if *sshCmd != "" && (*sshScpSourceFile != "" || *sshScpDestPath != "") {
		log.Fatal("请先传可执行文件到远程主机,之后再执行,为了并发安全,暂不支持传文件和执行命令同时运行")
	} else if *sshCmd != "" {
		fmt.Println(utils.Green("远程主机开始批量执行命令..................."))
	} else if *sshScpSourceFile != "" && *sshScpDestPath != "" {
		fmt.Println(utils.Green("开始传输文件到远程主机..................."))
	}

	//协程计数使用
	var m sync.Mutex

	//打开脚本运行成功日志文件
	logFile := open_log("task.log")
	//打开脚本运行错误日志文件
	errLogFile := open_log("task_error.log")
	//打开脚本运行错误日志文件
	ipStillRunningFile := open_log("task_ipstillrunning.log")

	utils.Write_log(logFile,"############### 新的并发任务开始  #################")
	utils.Write_log(errLogFile,"############### 新的并发任务开始  #################")
	utils.Write_log(ipStillRunningFile,"############### 新的并发任务开始  #################")

	defer logFile.Close()
	defer errLogFile.Close()
	defer ipStillRunningFile.Close()



	line_number := utils.ParseIpFile("ip.txt",file_map_ips,file_map_port,file_map_user,file_map_password,file_map_keyfilepassword)


	//先计算当前ip的总数量
	len_file_map_ips := len(file_map_ips)


	//在这里开启ssh协程,从下标1开始
	for i:=1; i<=len_file_map_ips; i++ {
		//传ip端口,用户名,密码到协程
		go ssh_command(i,file_map_ips[i],logFile,errLogFile,&m)
	}

	//检查执行结果协程
	go func() {
		for {
			time.Sleep(1 * time.Second)

			fmt.Println("运行成功的主机ip数: " + strconv.Itoa(runSuccessNumber))

		}
	}()


	//如果ip多,运行时间可能会比较长,每10秒需要检查一下当前有多少ip还在执行中
	go func() {
		for {
			time.Sleep(10 * time.Second)
			ipRunningCount := len(file_map_ips)


			fmt.Println(utils.Red("当前还在运行的ip总数: " + strconv.Itoa(ipRunningCount) ))
			utils.Write_log(ipStillRunningFile,"当前还在运行的ip总数: " + strconv.Itoa(ipRunningCount) )

			for _,ip :=range file_map_ips{
				fmt.Println(utils.Red(ip))

				utils.Write_log(ipStillRunningFile,ip)
			}

		}
	}()

	//在这里阻塞,等待每个协程返回的结果,从下标1开始
	for i:=1; i<=len_file_map_ips; i++ {
		//将结果写入日志
		//这里判断一下是不是包含成功的字符,写入不同的日志文件

		resultInfo = <-ipResultChanel

		if strings.Contains(resultInfo, "password") {
			fmt.Println("请检查用户名和密码有没有错误")
		//写入成功信息
		} else if strings.Contains(resultInfo,"成功"){
			utils.Write_log(logFile,resultInfo)
		//写入错误信息,空白不写入
		} else if !strings.Contains(resultInfo,""){
			utils.Write_log(errLogFile,resultInfo)
		}

	}

	endTime := time.Now()



	fmt.Printf("多ssh执行完毕. 处理时间是 %s \n", endTime.Sub(startTime))
	fmt.Println("运行结束")


	fmt.Println(utils.Green("运行成功的主机ip数: " + strconv.Itoa(runSuccessNumber)))
	fmt.Println(utils.Red("运行失败的主机ip数: " + strconv.Itoa(line_number - runSuccessNumber)))

}

/**
	index  		当前协程标识
	ip     		ip地址
	logFile 	日志文件
	errLogFile	错误日志
	m			同步锁
 */
func ssh_command(index int,ip string,logFile *os.File,errLogFile *os.File,m *sync.Mutex)  {

	var err error
	var signer ssh.Signer

	//认证方法
	var authMethod = []ssh.AuthMethod{}

	//如果密码是key,先认为是证书文件,密码只有3位通常也不认为是安全密码
	if file_map_password[index] == "key" {

		////读取证书文件
		privateKey, err := ioutil.ReadFile("id_rsa")
		if err != nil {

			goSshError(errLogFile,ip,"读取证书错误: "+ err.Error(),m,index)

			return
		}

		if file_map_keyfilepassword[index] == "" {
			//没有密码的证书
			signer, err = ssh.ParsePrivateKey(privateKey)
		} else {

			// 去除空格
			file_map_keyfilepassword[index] = strings.Replace(file_map_keyfilepassword[index], " ", "", -1)
			// 去除换行符
			file_map_keyfilepassword[index] = strings.Replace(file_map_keyfilepassword[index], "\n", "", -1)

			signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey,[]byte(file_map_keyfilepassword[index] + "dasf"))
		}

		if err != nil {

			goSshError(errLogFile,ip,"证书解析有问题,请检查证书文件是否正确" + err.Error(),m,index)

			return
		}

		//使用证书验证方法
		authMethod = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}

	} else {
		//使用密码验证方法
		authMethod = []ssh.AuthMethod{
			ssh.Password(file_map_password[index]),
		}
	}

	config1 := &ssh.ClientConfig{
		User: file_map_user[index],
		Auth: authMethod,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 60 * time.Second,
	}

	//ip加上端口
	client, err := ssh.Dial("tcp", file_map_ips[index]+":"+file_map_port[index], config1)

	if err != nil {

		goSshError(errLogFile,ip," 连接错误详细信息: " + err.Error(),m,index)

		return
	}




	//如果参数中包含处理sftp传输文件
	sftp, err := sftp.NewClient(client)
	if err != nil {

		goSshError(errLogFile,ip,"创建sftp传输链接错误: "+ err.Error(),m,index)

		return
	}
	defer sftp.Close()





	//判断需不需要传输远程脚本文件
	if *sshScpSourceFile != "" && *sshScpDestPath != "" {



		//先判断有没有斜线结果,组装成一个正常的远程路径
		//先去掉字符最尾的空白
		*sshScpDestPath = strings.Replace(*sshScpDestPath, " ", "", -1)

		//是否匹配字符串
		// .匹配任意一个字符 ，*匹配零个或多个 ，优先匹配更多(贪婪)
		match, _ := regexp.MatchString("/$", *sshScpDestPath)




		//先判断是不是/根路径
		if *sshScpDestPath == "/" {
			// /开头的路径
		} else if match {
			// 删除以/结尾的最后一个字符
			*sshScpDestPath = strings.TrimRight(*sshScpDestPath,"/")
		}

		remotePath := *sshScpDestPath + "/" + *sshScpSourceFile

		//组成一个完整的远程路径
		f, err := sftp.Create(remotePath)
		if err != nil {

			goSshError(errLogFile,ip,"创建远程文件错误: "+ err.Error(),m,index)

			return
		}


		//先去掉字符最尾的空白
		*sshScpSourceFile = strings.Replace(*sshScpSourceFile, " ", "", -1)
		scpfiledata,err := ioutil.ReadFile(*sshScpSourceFile)

		if err != nil {

			goSshError(errLogFile,ip,"读取本地传输文件出错: "+ err.Error(),m,index)

			return
		}

		if _, err := f.Write(scpfiledata); err != nil {

			goSshError(errLogFile,ip,"写入远程文件出错: "+ err.Error(),m,index)

			return
		}


		//给远程文件赋可执行权限
		err = sftp.Chmod(remotePath,0755)

		if err != nil {
			goSshError(errLogFile,ip,"给远程主机可执行文件权限出错: "+ err.Error(),m,index)
		}

		//在这里要将结果返回给chanel,这样才能将程序退出
		ipResultChanel<- ip + "传输脚本到远程主机成功"

		//删除当前正在运行的ip
		m.Lock()
		runSuccessNumber += 1
		delete(file_map_ips, index)
		defer m.Unlock()

	} else {

		//这个分支是执行命令
		//建立ssh连接
		session, err := client.NewSession()
		if err != nil {

			goSshError(errLogFile,ip,"创建ssh链接错误: "+ err.Error(),m,index)

			return
		}
		defer session.Close()


		//执行脚本,获取标准输入和错误信息
		var sshStdout bytes.Buffer
		var sshStderr bytes.Buffer

		session.Stdout = &sshStdout
		session.Stderr = &sshStderr


		if err := session.Run(*sshCmd); err != nil {

			goSshError(errLogFile,ip," 执行脚本错误: " + sshStderr.String(),m,index)

			return
		} else {
			//执行成功写入对应的日志文件
			m.Lock()
			runSuccessNumber += 1

			//删除当前正在运行的ip
			delete(file_map_ips, index)
			defer m.Unlock()

			//在这里要将结果返回给chanel,这样才能将程序退出
			ipResultChanel<-"执行脚本成功: " + ip + "\n" + sshStdout.String()

		}




	}








}



//打开日志文件
func open_log(log_name string) *os.File {
	// 定义一个文件
	fileName := log_name

	//没有的话就打开这个文件
	logFile,err  := os.OpenFile(fileName,os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)

	if err != nil {
		log.Fatalln("打开日志文件错误 !")
	}

	return logFile

}


//协程中出现错误处理函数
func goSshError(errLogFile *os.File,ip string,errorInfo string,m *sync.Mutex,index int) {
	//写入到并发错误日志文件
	utils.Write_log(errLogFile,ip + " " + errorInfo)

	//写入日志后,返回空字符串即可
	ipResultChanel<-""

	//删除当前正在运行的ip
	m.Lock()
	delete(file_map_ips, index)
	defer m.Unlock()
}






