module github.com/gomanyssh

go 1.14

require (
	github.com/pkg/sftp v1.11.0
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586
)

//go mod可以引用自己,将 github.com/gomanyssh 改为./方式
//使用import "../fetch"  不是很好看
replace github.com/gomanyssh => ./
