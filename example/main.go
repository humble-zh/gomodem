package main

import (
	"fmt"
	"sync"
	"time"

	modem "github.com/humble-zh/gomodem"
)

var (
	wg sync.WaitGroup
)

func main() {
	mtot := 1
	mslice := make([]modem.IModem, mtot)
	// fmt.Printf("%#v\n", mslice)

	// m, err := modem.NewWithJsonBytes([]byte("{\"model\":\"ep06\",\"findifacename\":\"ls -l /sys/class/net |awk -F'[/]' '{if($9~/1-1:1.4/){ print $NF }}'\",\"findatdevpath\":\"ls -l /sys/class/tty/ttyUSB*|awk -F'[/]' '{if($13~/1-1:1.3/){ print \\\"/dev/\\\"$NF }}'\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	m, err := modem.NewWithJsonBytes([]byte("{\"model\":\"ep06\",\"usbbusportid\":\"1-1:1\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	if err != nil {
		fmt.Println(err)
		return
	}
	mslice[0] = m
	// m, err = modem.NewWithJsonBytes([]byte("{\"model\":\"rm500q\",\"findifacename\":\"ohno\",\"findatdevpath\":\"haha\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// mslice[1] = m

	for i := 0; i < mtot; i++ {
		m = mslice[i]
		if err = m.Open(); err != nil {
			fmt.Println(err)
			return
		}
		defer m.Close()
		fmt.Printf("%+v\n", m)

		wg.Add(1)
		go modem.Start(m, &wg)
	}

	time.Sleep(time.Second * 60)

	for i := 0; i < mtot; i++ {
		modem.Stop(mslice[i])
	}
	wg.Wait()
}
