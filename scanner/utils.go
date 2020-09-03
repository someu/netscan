package scanner

import (
	"crypto/md5"
	"fmt"
)

func md5v(input string) string {
	data := []byte(input)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}
