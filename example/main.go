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
	mtot := 2
	mslice := make([]modem.IModem, mtot)
	// fmt.Printf("%#v\n", mslice)

	m, err := modem.NewWithJsonBytes([]byte("{\"model\":\"ep06\",\"findifacename\":\"hello\",\"findatdevpath\":\"world\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	if err != nil {
		fmt.Println(err)
		return
	}
	mslice[0] = m
	m, err = modem.NewWithJsonBytes([]byte("{\"model\":\"rm500q\",\"findifacename\":\"ohno\",\"findatdevpath\":\"haha\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	if err != nil {
		fmt.Println(err)
		return
	}
	mslice[1] = m
	// fmt.Printf("%+v\n", mslice)

	for i := 0; i < mtot; i++ {
		m = mslice[i]
		if err = m.Open(); err != nil {
			fmt.Println(err)
			return
		}
		defer m.Close()

		wg.Add(1)
		go modem.Start(m, &wg)
	}

	time.Sleep(time.Second * 5)
	for i := 0; i < mtot; i++ {
		modem.Stop(mslice[i])
	}
	wg.Wait()
}
