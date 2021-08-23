package kernel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/alibabacloud-go/tea/tea"
)

func Log(val *string) {
	fmt.Printf("[LOG] %s\n", tea.StringValue(val))
}

func Sha256(message string) string {
	bytes2:=sha256.Sum256( []byte(message))//计算哈希值，返回一个长度为32的数组
	hashcode2:=hex.EncodeToString(bytes2[:])//将数组转换成切片，转换成16进制，返回字符串
	return hashcode2
}