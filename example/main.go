package main

import (
	"fmt"

	modem "github.com/humble-zh/gomodem"
)

func main() {
	m, err := modem.NewWithJsonBytes([]byte("{\"model\":\"ep06\",\"findifacename\":\"hello\",\"findatdevpath\":\"world\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	// m, err := modem.NewWithJsonBytes([]byte("{\"model\":\"rm500q\",\"findifacename\":\"ohno\",\"findatdevpath\":\"haha\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	if err != nil {
		fmt.Println(err)
		return
	}
	if m.IsOK() == nil {
		fmt.Println("ok")
	}
	if err = m.Open(); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", m)
}
