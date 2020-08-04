package confutil

/*
func TestMapTo1(t *testing.T) {
	type A struct {
		Max  int     `ini:"max"`
		Port string  `ini:"port"`
		Rate float32 `ini:"rate"`
	}
	var a A
	data := make(map[string]string, 0)
	data["max"] = "101"
	data["port"] = ":4001"
	data["rate"] = "0.03"
	e := TestMapTo(data, &a)
	fmt.Println(e)
	fmt.Printf("get a:[%+v]\n", a)
}

func TestMapTo2(t *testing.T) {
	type A struct {
		Max  []int     `ini:"max"`
		Port []string  `ini:"port"`
		Rate []float32 `ini:"rate"`
	}
	var a A
	data := make(map[string]string, 0)
	data["max"] = "101 102 103"
	data["port"] = ":4001 9i99i rrf3"
	data["rate"] = "0.03 1.23 5.34"
	e := TestMapTo(data, &a)
	fmt.Println(e)
	fmt.Printf("get a:[%+v]\n", a)
}

func TestMapTo3(t *testing.T) {
	type A struct {
		MaxTimeout  time.Duration `ini:"max"`
		ReleaseTime time.Time     `ini:"port"`
	}
	var a A
	data := make(map[string]string, 0)
	data["max"] = "101s"
	data["port"] = "2019-08-22 23:04:05"
	e := TestMapTo(data, &a)
	fmt.Println(e)
	fmt.Printf("get a:[%+v]\n", a)
	fmt.Println(a.ReleaseTime.Unix())
}

func TestMapTo4(t *testing.T) {
	type A struct {
		Max  int `ini:"max"`
		Port string
		Rate float32 `ini:"rate"`
	}
	var a A
	data := make(map[string]string, 0)
	data["max"] = "101"
	data["port"] = ":4001"
	data["rat"] = "0.03"
	e := TestMapTo(data, a)
	fmt.Println(e)
	e2 := TestMapTo(data, &a)
	fmt.Println(e2)
	fmt.Printf("get a:[%+v]\n", a)
}
*/
