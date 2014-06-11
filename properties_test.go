package properties

import (
	"io/ioutil"
	"os"
	"testing"
)

var (
	ValidConfiguration = `
# test configuration file

source = some source
destination = some destination # with comment
bind_port = 9345

[section1]
float = 4.5
`
	InvalidConfiguration = `
# test configuration
this is an invalid line
this_is = valid one
`
)

func tmpConfigFile(data string) (string, error) {
	tmp, err := ioutil.TempFile(".", "config_")
	if err != nil {
		return "", err
	}
	_, err = tmp.Write([]byte(data))
	if err != nil {
		return "", err
	}
	tmp.Close()
	return tmp.Name(), nil

}

func TestParse1(t *testing.T) {
	filename, err := tmpConfigFile(ValidConfiguration)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(filename)

	props, err := Load(filename)
	if err != nil {
		t.Error(err)
		return
	}

	// reading string properties

	source, err := props.GetString("source")
	if err != nil {
		t.Error(err)
	}
	if source != "some source" {
		t.Errorf("Invalid value: got %s, expected %s", source, "some source")
	}

	destination, err := props.GetString("destination")
	if err != nil {
		t.Error(err)
	}
	if destination != "some destination" {
		t.Errorf("Invalid value: got %s, expected %s", destination, "some destination")
	}

	// reading non-existent property
	_, err = props.GetString("non_existent")
	if err != NodeNotFound {
		t.Error("Something wrong happened: non_existent property is found")
	}

	// reading existent node with no value
	_, err = props.GetString("section1")
	if err == nil {
		t.Error("Something wrong happened: node 'section1' must return NoValueError")
	}

	// reading integer property
	port, err := props.GetInt("bind_port")
	if err != nil {
		t.Error(err)
	}
	if port != 9345 {
		t.Errorf("Invalid integer value of 'port': got %s, expected %s", port, 9345)
	}

	// reading sectioned float property
	flvalue, err := props.GetFloat("section1.float")
	if err != nil {
		t.Error(err)
	}
	if flvalue != 4.5 {
		t.Errorf("Invalid float value of 'section1.float': got %s, expected %s", flvalue, 4.5)
	}

}

func TestParse2(t *testing.T) {
	filename, err := tmpConfigFile(InvalidConfiguration)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(filename)

	_, err = Load(filename)
	if err == nil {
		t.Error("Parsing invalid configuration with no error!")
		return
	}
}

func TestKeyExist(t *testing.T) {
	filename, err := tmpConfigFile(ValidConfiguration)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(filename)

	props, err := Load(filename)
	if err != nil {
		t.Error(err)
		return
	}

	if !props.KeyExists("section1") {
		t.Error("KeyExists('section1') returns false")
	}

	if props.KeyExists("non_existent") {
		t.Error("KeyExists('non_existent') returns true")
	}
}
