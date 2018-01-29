package parse

import (
	"io/ioutil"
	"log"

	"spidermesh/config"

	"gopkg.in/yaml.v2"
)

func Parse(f string) *config.Config {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		log.Printf("Fail to read file %s, %s", f, err.Error())
		return nil
	}
	t := config.Config{}
	err = yaml.Unmarshal(data, &t)
	if err != nil {
		log.Printf("Fail to unmarshal data, %s", err.Error())
		return nil
	}
	return &t
}
