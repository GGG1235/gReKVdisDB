package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type JsonToMap interface {
	OpenFile(string) ([]byte, error)
	ToMap(string) (map[string]interface{},error)
}

type JsonObject struct {}

func NewJsonObject() *JsonObject {
	return &JsonObject{}
}

func (jsonObject *JsonObject) OpenFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil,err
	}
	return ioutil.ReadAll(file)
}

func (jsonObject *JsonObject) ToMap(path string) (map[string]interface{},error) {
	f,err := jsonObject.OpenFile(path)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	content := make(map[string]interface{})
	err = json.Unmarshal([]byte(f),&content)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	return content, nil
}