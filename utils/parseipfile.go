package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func ParseIpFile(ipFileName string, ips map[int]string, port map[int]string, user map[int]string, password map[int]string, keyfilepassword map[int]string) int {

	//读取ip和密码/证书文件
	fileEveryLine, err := os.Open(ipFileName)
	if err != nil {
		log.Fatal("读取ip文件错误: \n", err)
	}
	defer fileEveryLine.Close()

	br := bufio.NewReader(fileEveryLine)

	//当前文件所在的行号
	lineNumber := 0

	for {
		//行号加一
		lineNumber++

		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}


		spaceRe, _ := regexp.Compile(`\s+`)

		s := spaceRe.Split(string(line), -1)

		//ip
		if !checkIp(s[0]) {
			log.Fatal("第" + strconv.Itoa(lineNumber) + "行 : 获取ip出错,请检查ip格式有没有填写正确,填写ip的前面不能有空格,文件不能包括空行")
		} else  {
			ips[lineNumber] = s[0]
		}

		//端口
		if !checkPort(s[1]) {
			log.Fatal("第" + strconv.Itoa(lineNumber) + "行 : 获取端口出错,请检查端口格式有没有填写正确,1到65535")
		} else {
			port[lineNumber] = s[1]
		}

		//用户名
		if s[2] == "" {
			log.Fatal("第" + strconv.Itoa(lineNumber) + "行 : 获取用户名出错,请检查用户名有没有填写正确")
		} else {
			user[lineNumber] = s[2]
		}

		//密码或证书文件
		if s[3] == "" {
			log.Fatal("第" + strconv.Itoa(lineNumber) + "行 : 密码为空或没有证书文件,请检查密码或证书有没有填写正确")
		} else {
			password[lineNumber] = s[3]
		}

		//如果证书有密码的,需要填写证书密码,如果没有的话默认是空密码

		//注意先后顺序,先有第5个字段,才判断是否为空
		//直接取第5个字段会有可能超出下标
		if len(s) >= 5 {
			if s[4] != "" {
				keyfilepassword[lineNumber] = s[4]
			}
		}

	}


	//前面行号减1
	lineIpCount := lineNumber - 1
	fmt.Println("总ip数是:" + strconv.Itoa(lineIpCount))

	return lineIpCount
}


//检查ip地址是否正确
func checkIp(ip string) bool {
	addr := strings.Trim(ip, " ")
	regStr := `^(([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.)(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){2}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
	if match, _ := regexp.MatchString(regStr, addr); match {
		return true
	}
	return false
}

func checkPort(port string) bool {

	regStr := `^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`
	if match, _ := regexp.MatchString(regStr, port); match {
		return true
	}

	return false
}


