package util

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"strings"
)

func Base64Decode(in string) (string, error) {
	if m := len(in) % 4; m != 0 {
		in += strings.Repeat("=", 4-m)
	}
	r := bytes.NewReader([]byte(in))
	// pass it to NewDecoder so that it can read data
	dec := base64.NewDecoder(base64.StdEncoding, r)
	// read decoded data from dec to res
	res, err := ioutil.ReadAll(dec)
	return string(res), err
}
