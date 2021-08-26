package kernel

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	Sort "sort"
	"strconv"
	"strings"
	"time"
)

/**
获取时间戳，格式yyyy-MM-dd HH:mm:ss
*/

var this Client


func InitClient(config Client) {
	this = config
}

/**
解析网关响应内容，同时将API的接口名称和响应原文插入到响应数组的method和body字段中
*/
func ReadAsJson(response io.Reader) (map[string]string, error) {
	byt, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	err = json.Unmarshal(byt, &m)
	return m,err
}


func GetConfig(key string) string {
	if key == "protocol" {
		return this.Protocol
	} else if key == "gatewayHost" {
		return this.GatewayHost
	} else if key == "appId" {
		return this.AppId
	} else if key == "merchantId" {
		return this.MerchantId
	} else if key == "secretKey" {
		return this.SecretKey
	} else if key == "accessToken" {
		return this.AccessToken
	}else {
		panic(key + " is illegal")
	}
}

func GetTimestamp() string {
	return strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
}

/**
将业务参数和其他额外文本参数按www-form-urlencoded格式转换成HTTP Body中的字节数组，注意要做URL Encode
*/
func ToUrlEncodedRequestBody(bizParams map[string]string) string {
	sortedMap := GetSortedMap(nil, bizParams, nil)
	if sortedMap == nil {
		return ""
	}
	return buildQueryString(bizParams)
}

func MergeMap(mObj ...map[string]string) map[string]string {
	newObj := map[string]string{}
	for _, m := range mObj {
		for k, v := range m {
			newObj[k] = v
		}
	}
	return newObj
}

func ToRespModel(resp map[string]string) map[string]interface{} {
	code := resp[code]
	msg := resp[msg]
	m := make(map[string]interface{})
	if len(code) > 0 && code != success_code {
		data := resp[biz_content_field]
		if len(data) > 0 {
			err := json.Unmarshal([]byte(data), &m)
			if err == nil {
				return m
			}
		}
	}
	panic("接口访问异常，code:" +code+",msg:" + msg)

}

func toJSONString(params map[string]interface{}) string {
	mjson,_ :=json.Marshal(params)
	mString :=string(mjson)
	return mString
}

func ObjToJSONString(model interface{}) string{
	marshal, err := json.Marshal(model)
	if err != nil {
		return ""
	}
	mString :=string(marshal)
	return mString
}

func sortMap(params map[string]string) map[string]string {
	return params;
}


func Sign(systemParams map[string]string, bizParams map[string]string, textParams map[string]string,secretKey string) string {
	sortedMap := GetSortedMap(systemParams, bizParams, textParams)
	//var content = "";
	var index = 0;
	var _content string;
	for key, value := range sortedMap {
		if len(key)>0 && len(value) > 0{
			var temp = "&";
			if index == 0{
				temp = "";
			}else {
				temp = "&";
			}
			_content = _content + temp + key + "=" + value
			index++;
		}
	}
	newStr := secretKey + _content
	return Sha256(newStr)
}

func GetSortedMap(systemParams map[string]string, bizParams map[string]string, textParams map[string]string) map[string]string {
	sortedMap := MergeMap(systemParams, bizParams, textParams)
	newMap := map[string]string{}
	EachMap(sortedMap, func(key string, value string){
		newMap[key] = value
	})
	return newMap
}

// 以map的key(int\float\string)排序遍历map
// eachMap      ->  待遍历的map
// eachFunc     ->  map遍历接收，入参应该符合map的key和value
// 需要对传入类型进行检查，不符合则直接panic提醒进行代码调整
func EachMap(eachMap interface{}, eachFunc interface{})  {
	eachMapValue := reflect.ValueOf(eachMap)
	eachFuncValue := reflect.ValueOf(eachFunc)
	eachMapType := eachMapValue.Type()
	eachFuncType := eachFuncValue.Type()
	if eachMapValue.Kind() != reflect.Map {
		panic(errors.New("ksort.EachMap failed. parameter \"eachMap\" type must is map[...]...{}"))
	}
	if eachFuncValue.Kind() != reflect.Func {
		panic(errors.New("ksort.EachMap failed. parameter \"eachFunc\" type must is func(key ..., value ...)"))
	}
	if eachFuncType.NumIn() != 2 {
		panic(errors.New("ksort.EachMap failed. \"eachFunc\" input parameter count must is 2"))
	}
	if eachFuncType.In(0).Kind() != eachMapType.Key().Kind() {
		panic(errors.New("ksort.EachMap failed. \"eachFunc\" input parameter 1 type not equal of \"eachMap\" key"))
	}
	if eachFuncType.In(1).Kind() != eachMapType.Elem().Kind() {
		panic(errors.New("ksort.EachMap failed. \"eachFunc\" input parameter 2 type not equal of \"eachMap\" value"))
	}

	// 对key进行排序
	// 获取排序后map的key和value，作为参数调用eachFunc即可
	switch eachMapType.Key().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		keys := make([]int, 0)
		keysMap := map[int]reflect.Value{}
		for _, value := range eachMapValue.MapKeys() {
			keys = append(keys, int(value.Int()))
			keysMap[int(value.Int())] = value
		}
		Sort.Ints(keys)
		for _, key := range keys {
			eachFuncValue.Call([]reflect.Value{keysMap[key], eachMapValue.MapIndex(keysMap[key])})
		}
	case reflect.Float64, reflect.Float32:
		keys := make([]float64, 0)
		keysMap := map[float64]reflect.Value{}
		for _, value := range eachMapValue.MapKeys() {
			keys = append(keys, float64(value.Float()))
			keysMap[float64(value.Float())] = value
		}
		Sort.Float64s(keys)
		for _, key := range keys {
			eachFuncValue.Call([]reflect.Value{keysMap[key], eachMapValue.MapIndex(keysMap[key])})
		}
	case reflect.String:
		keys := make([]string, 0)
		keysMap := map[string]reflect.Value{}
		for _, value := range eachMapValue.MapKeys() {
			keys = append(keys, value.String())
			keysMap[value.String()] = value
		}
		Sort.Strings(keys)
		for _, key := range keys {
			eachFuncValue.Call([]reflect.Value{keysMap[key], eachMapValue.MapIndex(keysMap[key])})
		}
	default:
		panic(errors.New("\"eachMap\" key type must is int or float or string"))
	}
}

func SortMap(romanNumeralDict map[string]string) map[string]string {
	keys := make([]string, 0)
	for k, _ := range romanNumeralDict {
		keys = append(keys, k)
	}
	Sort.Strings(keys)
	return romanNumeralDict
}

func buildQueryString(sortedMap map[string]string) string {
	requestUrl := ""
	keys := make([]string, 0)
	for k, _ := range sortedMap {
		keys = append(keys, k)
	}
	Sort.Strings(keys)
	var pList = make([]string, 0, 0)
	for _, key := range keys {
		pList = append(pList, key+"="+sortedMap[key])
	}
	requestUrl = strings.Join(pList, "&")
	return requestUrl
}



type Client struct {
	Protocol           string
	GatewayHost        string
	AppId              string
	MerchantId         string
	SecretKey    	   string
	AccessToken 	   string
}
