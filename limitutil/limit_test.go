package limitutil

import "testing"
import "syscall"

func Test_GrowToMaxFdLimit_AfterCall_CurMaxEquals(t *testing.T) {
	//	make cur and max fd not equal
	max, err := GetMaxFdLimit()
	if err != nil {
		t.Fatal("Unexpected error on testing")
	}
	t.Log(max, "===maxx")

	if err = SetCurFdLimit(max - 1); err != nil {
		t.Fatal("Unexpected error on testing")
	}

	///////
	if err := GrowToMaxFdLimit(); err != nil {
		t.Fatal(err)
	}
	///////

	cur, errr := GetCurFdLimit()
	if errr != nil {
		t.Fatal("Unexpected error on testing")
	}
	max, err = GetMaxFdLimit()
	if err != nil {
		t.Fatal("Unexpected error on testing")
	}

	if cur != max {
		t.Error("Not equal:", cur, max)
	}
}

////////////////////////////////////////////////////////////////////////////////
//	WARNING:	Run this test under root account !!
func Test_GrowToMaxFdRoot_MaxLessThanMaxRoot_CurEqualsMaxRoot(t *testing.T) {
	var rlimit syscall.Rlimit
	rlimit.Cur = 1000 // this value less than the max limit under root privilege
	rlimit.Max = 1000

	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		t.Fatal("Unexpected error on testing")
	}

	///////
	if err := GrowToMaxFdLimitRoot(); err != nil {
		t.Fatal(err)
	}
	//////

	expectedMax, er := GetMaxFdLimitRoot()
	if er != nil {
		t.Fatal("Unexpected error on testing")
	}
	actualMax, errr := GetCurFdLimit()
	if errr != nil {
		t.Fatal("Unexpected error on testing")
	}

	if expectedMax != actualMax {
		t.Error("Not equal:", expectedMax, actualMax)
	}
}
