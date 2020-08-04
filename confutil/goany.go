package confutil

import (
	"strings"

	"github.com/spf13/cast"
	yaml "gopkg.in/yaml.v2"
)

//any struct
//example for config plugin
//no config source type limited
//only need to write the load function to init the data
type AnyFile struct {
	data map[interface{}]interface{} // Section -> key : value
}

//set value
//no need clear the cache after set value,because it get value form the map instead of load file
func (this *AnyFile) Set(section, key string, value interface{}) {
	smap := make(map[interface{}]interface{}, 0)
	smap[key] = value
	this.data[section] = smap
}

//load function
func loadAny() (cfg Config, err error) {
	anyFile := new(AnyFile)
	//example for init data
	amap := make(map[interface{}]interface{}, 1)
	amap["Redis"] = map[interface{}]interface{}{"redis": "127.0.0.1:6379 127.0.0.1:7379"}
	anyFile.data = amap
	return anyFile, nil
}

//MustValue implemented
func (this *AnyFile) MustValue(section, key string, defaultVal ...string) string {
	defaultValue := ""
	if len(defaultVal) > 0 {
		defaultValue = defaultVal[0]
	}

	if val, ok := this.data[section]; !ok {
		return defaultValue
	} else {
		if data, ok := val.(map[interface{}]interface{}); ok {
			for k, v := range data {
				if cast.ToString(k) == key {
					return cast.ToString(v)
				}
			}
		} else {
			return defaultValue
		}
	}
	return defaultValue
}

//MustValueArray implemented
func (this *AnyFile) MustValueArray(section, key, delim string) []string {
	val := this.MustValue(section, key, "")
	if val != "" {
		return strings.Split(val, delim)
	}
	return nil
}

//GetKeyList implemented
func (this *AnyFile) GetKeyList(section string) []string {
	if val, err := this.GetSection(section); err != nil {
		return nil
	} else {
		data := make([]string, len(val))
		for k, _ := range val {
			data = append(data, k)
		}
		return data
	}
	return nil

}

//empty function
func (this *AnyFile) GetSectionList() []string {
	return nil
}

//GetSection implemented
func (this *AnyFile) GetSection(section string) (map[string]string, error) {
	if val, ok := this.data[section]; !ok {
		return nil, nil
	} else {
		if data, ok := val.(map[interface{}]interface{}); ok {
			ret := make(map[string]string, len(data))
			for k, v := range data {
				ret[cast.ToString(k)] = cast.ToString(v)
			}
			return ret, nil
		}
	}
	return nil, nil
}

//GetSectionObject implemented
func (this *AnyFile) GetSectionObject(section string, obj interface{}) error {
	if section == "" {
		byt, err := yaml.Marshal(this.data)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(byt, obj)
		if err != nil {
			return err
		}
	} else if val, ok := this.data[section]; !ok {
		return nil
	} else {
		byt, err := yaml.Marshal(val)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(byt, obj)
		if err != nil {
			return err
		}
	}
	return nil
}
