/*===============================================================
*   Copyright (C) 2019 All rights reserved.
*
*   FileName：goconfig.go
*   Author：WuGuoFu
*   Date： 2019-12-04
*   Description：
*
================================================================*/
package confutil

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Unknwon/goconfig"
)

//ini struct
//implement the Config interface
type IniFile struct {
	*goconfig.ConfigFile
}

func (this *IniFile) Set(section, key string, value interface{}) {
	//not support
}

//load function
func loadIniFile(path string) (cfg Config, err error) {
	//load file
	file, err := goconfig.LoadConfigFile(path)
	config := new(IniFile)
	config.ConfigFile = file
	return config, err
}

//GetSectionObject implemented
//obj must a pointer
func (ini *IniFile) GetSectionObject(section string, obj interface{}) error {
	if ret, err := g_cfg.GetSection(section); err != nil {
		log.Printf("Conf,err:%v", err)
		return err
	} else {
		//reflect
		typ := reflect.TypeOf(obj)
		val := reflect.ValueOf(obj)
		if typ.Kind() == reflect.Ptr {
			val = val.Elem()
		} else {
			return errors.New("cannot map to non-pointer struct")
		}
		return mapTo(ret, val)
	}
}

//convert map to object
func mapTo(info map[string]string, val reflect.Value) error {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := val.Field(i)
		tpField := typ.Field(i)
		tag := tpField.Tag.Get("ini")
		if tag == "-" || !field.CanSet() {
			continue
		}
		if value, ok := info[tag]; ok {
			if err := setWithProperType(tpField.Type, value, field); err != nil {
				return fmt.Errorf("error mapping field(%s): %v", tag, err)
			}
		}
	}
	return nil
}

var reflectTime = reflect.TypeOf(time.Now()).Kind()

//format object value
func setWithProperType(t reflect.Type, value string, field reflect.Value) error {
	switch t.Kind() {
	case reflect.String:
		if len(value) == 0 {
			return nil
		}
		field.SetString(value)
	case reflect.Bool:
		boolVal, _ := strconv.ParseBool(value)
		field.SetBool(boolVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		durationVal, err := time.ParseDuration(value)
		// Skip zero value
		if err == nil && int64(durationVal) > 0 {
			field.Set(reflect.ValueOf(durationVal))
			return nil
		}
		intVal, _ := strconv.ParseInt(value, 10, 64)
		field.SetInt(intVal)
	case reflect.Float32, reflect.Float64:
		floatVal, _ := strconv.ParseFloat(value, 64)
		field.SetFloat(floatVal)
	case reflectTime:
		location, _ := time.LoadLocation("Asia/Shanghai")
		timeVal, err := time.ParseInLocation("2006-01-02 15:04:05", value, location)
		if err != nil {
			return fmt.Errorf("time parse error: %s", value)
		}
		field.Set(reflect.ValueOf(timeVal))
	case reflect.Slice:
		var strs []string
		strs = strings.Split(value, " ")
		numVals := len(strs)
		if numVals == 0 {
			return nil
		}
		sliceOf := field.Type().Elem().Kind()
		slice := reflect.MakeSlice(field.Type(), numVals, numVals)
		for i := 0; i < numVals; i++ {
			switch sliceOf {
			case reflect.String:
				slice.Index(i).Set(reflect.ValueOf(strs[i]))
			case reflect.Int:
				intVal, _ := strconv.Atoi(strs[i])
				slice.Index(i).Set(reflect.ValueOf(intVal))
			case reflect.Int64:
				int64Val, _ := strconv.ParseInt(strs[i], 10, 64)
				slice.Index(i).Set(reflect.ValueOf(int64Val))
			case reflect.Uint:
				uint64Val, _ := strconv.ParseUint(strs[i], 10, 64)
				slice.Index(i).Set(reflect.ValueOf(uint(uint64Val)))
			case reflect.Uint64:
				uint64Val, _ := strconv.ParseUint(strs[i], 10, 64)
				slice.Index(i).Set(reflect.ValueOf(uint64Val))
			case reflect.Float64:
				float64Val, _ := strconv.ParseFloat(strs[i], 64)
				slice.Index(i).Set(reflect.ValueOf(float64Val))
			case reflect.Float32:
				float64Val, _ := strconv.ParseFloat(strs[i], 64)
				slice.Index(i).Set(reflect.ValueOf(float32(float64Val)))
			}
		}
		field.Set(slice)
	default:
		return fmt.Errorf("unsupported type '%s'", t)
	}
	return nil
}
