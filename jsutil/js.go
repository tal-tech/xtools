package jsutil

import (
	"fmt"
	"math"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"
)

//json instance with jsoniter
var Json = jsoniter.ConfigCompatibleWithStandardLibrary

func JsonMinify(str string) string {
	in_string := false
	in_single_comment := false
	in_multi_comment := false
	string_opener := string('x')

	var retStr string = ""

	size := len(str)
	for i := 0; i < size; i++ {

		c := fmt.Sprintf("%s", []byte{str[i]})
		next := math.Min(cast.ToFloat64(i+2), cast.ToFloat64(size))
		cc := fmt.Sprintf("%s", []byte(str[i:cast.ToInt64(next)]))

		if in_string {
			if c == string_opener {
				in_string = false
				retStr += c
			} else if c == "\\" {
				retStr += cc
				i++
			} else {
				retStr += c
			}
		} else if in_single_comment {
			if c == "\r" || c == "\n" {
				in_single_comment = false
			}
		} else if in_multi_comment {
			if cc == "*/" {
				in_multi_comment = false
				i++
			}
		} else {
			if cc == "/*" {
				in_multi_comment = true
				i++
			} else if cc == "//" {
				in_single_comment = true
				i++
			} else if c[0] == '"' || c[0] == '\'' {
				in_string = true
				string_opener = c
				retStr += c
			} else if c != " " && c != "\t" && c != "\n" && c != "\r" {
				retStr += c
			}
		}
	}
	return retStr
}
