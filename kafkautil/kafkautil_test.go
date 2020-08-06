/*===============================================================
*   Copyright (C) 2018 All rights reserved.
*
*   FileName：kafkautil_test.go
*   Author：WuGuoFu
*   Date： 2018-07-27
*   Description：
*
================================================================*/
package kafkautil

import (
	"fmt"
	"testing"
)

func TestSendProxy(t *testing.T) {
	fmt.Println(Send2Proxy("xes_submit_testv2", "xesv5 11 22 33 44 55 66 77"))
}
