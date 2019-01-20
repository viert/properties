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
otherkey = othervalue

bool.true = yes
bool.false = 0
bool.anotherfalse = false
bool.invalid = invalid

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
	if err != nodeNotFound {
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
		t.Errorf("Invalid integer value of 'port': got %d, expected %d", port, 9345)
	}

	// reading sectioned float property
	flvalue, err := props.GetFloat("section1.float")
	if err != nil {
		t.Error(err)
	}
	if flvalue != 4.5 {
		t.Errorf("Invalid float value of 'section1.float': got %f, expected %f", flvalue, 4.5)
	}

	var bv bool
	bv, err = props.GetBool("section1.bool.true")
	if err != nil {
		t.Error(err)
	}
	if !bv {
		t.Errorf("Invalid float value of 'section1.bool.true': got %v, expected %v", bv, true)
	}

	bv, err = props.GetBool("section1.bool.false")
	if err != nil {
		t.Error(err)
	}
	if bv {
		t.Errorf("Invalid float value of 'section1.bool.false': got %v, expected %v", bv, false)
	}

	bv, err = props.GetBool("section1.bool.anotherfalse")
	if err != nil {
		t.Error(err)
	}
	if bv {
		t.Errorf("Invalid float value of 'section1.bool.anotherfalse': got %v, expected %v", bv, false)
	}

	bv, err = props.GetBool("section1.bool.invalid")
	if err == nil {
		t.Error("Bool value section1.bool.invalid should be invalid")
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

func TestSubkeys(t *testing.T) {
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

	expected := make(map[string]bool)
	expected["float"] = true
	expected["otherkey"] = true
	expected["bool"] = true

	skeys, err := props.Subkeys("section1")
	if err != nil {
		t.Error(err)
		return
	}

	if len(skeys) != 3 {
		t.Error("Invalid number of subkeys, expected number is 3, found", len(skeys))
	}
	for _, key := range skeys {
		if _, ok := expected[key]; !ok {
			t.Errorf("Invalid key found '%s', expected ['float', 'otherkey']", key)
		}
	}

}

func TestSubkeysRoot(t *testing.T) {
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

	expected := make(map[string]bool)
	expected["source"] = true
	expected["destination"] = true
	expected["bind_port"] = true
	expected["section1"] = true

	skeys, err := props.Subkeys("")
	if err != nil {
		t.Error(err)
		return
	}

	if len(skeys) != 4 {
		t.Error("Invalid number of subkeys, expected number is 4, found", len(skeys))
	}
	for _, key := range skeys {
		if _, ok := expected[key]; !ok {
			t.Errorf("Invalid key found '%s'", key)
		}
	}

}
