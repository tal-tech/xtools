package confutil

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/cast"
	yaml "gopkg.in/yaml.v2"
)

//yaml struct
//support yaml file parse
//implemented Config interface
type YamlFile struct {
	data map[string]interface{} // Section -> key : value
}

//set function
//need clear cache after set value
func (this *YamlFile) Set(section, key string, value interface{}) {
	smap := make(map[interface{}]interface{}, 0)
	smap[key] = value
	this.data[section] = smap
}

//load yaml file
func loadYamlFile(path string) (cfg Config, err error) {
	//read file
	content, ioerr := ioutil.ReadFile(path)
	if ioerr != nil {
		err = ioerr
		log.Printf("loadYamlFile error: %v", ioerr)
		return nil, err
	}
	data := make(map[string]interface{}, 0)
	//convert bytes to map
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		log.Printf("loadYamlFile error: %v", err)
		return nil, err
	}
	yamlFile := new(YamlFile)
	yamlFile.data = data
	return yamlFile, nil
}

//MustValue function implemented
func (this *YamlFile) MustValue(section, key string, defaultVal ...string) string {
	defaultValue := ""
	if len(defaultVal) > 0 {
		defaultValue = defaultVal[0]
	}

	if val, ok := this.data[section]; !ok {
		return defaultValue
	} else {
		//match the key
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
//split value by delim,like ",","-"
func (this *YamlFile) MustValueArray(section, key, delim string) []string {
	val := this.MustValue(section, key, "")
	if val != "" {
		return strings.Split(val, delim)
	}
	return nil
}

//GetKeyList implemented
//get all keys
func (this *YamlFile) GetKeyList(section string) []string {
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
func (this *YamlFile) GetSectionList() []string {
	return nil
}

//GetSection implemented
func (this *YamlFile) GetSection(section string) (map[string]string, error) {
	if val, ok := this.data[section]; !ok {
		return nil, nil
	} else {
		//math the type
		if data, ok := val.(map[interface{}]interface{}); ok {
			ret := make(map[string]string, len(data))
			//format map key and value
			for k, v := range data {
				ret[cast.ToString(k)] = cast.ToString(v)
			}
			return ret, nil
		}
	}
	return nil, nil
}

//GetSectionObject implemented
//object must be a pointer
func (this *YamlFile) GetSectionObject(section string, obj interface{}) error {
	//no section,use all data
	if section == "" {
		//formate map to bytes
		byt, err := yaml.Marshal(this.data)
		if err != nil {
			return err
		}
		//format bytes to object
		err = yaml.Unmarshal(byt, obj)
		if err != nil {
			return err
		}
	} else if val, ok := this.data[section]; !ok {
		//not hit value
		return nil
	} else {
		//convert value to object by section
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
