package main

import (
	"sync"
	"time"

	modem "github.com/humble-zh/gomodem"
	"github.com/sirupsen/logrus"
)

var (
	wg sync.WaitGroup
)

func main() {
	log := logrus.New()
	log.SetReportCaller(true)
	// log.SetLevel(logrus.DebugLevel)

	mtot := 1
	mslice := make([]modem.IModem, mtot)
	// log.Debugf("%#v", mslice)

	// m, err := modem.NewWithJsonBytes([]byte("{\"model\":\"ep06\",\"findifacename\":\"ls -l /sys/class/net |awk -F'[/]' '{if($9~/1-1:1.4/){ print $NF }}'\",\"findatdevpath\":\"ls -l /sys/class/tty/ttyUSB*|awk -F'[/]' '{if($13~/1-1:1.3/){ print \\\"/dev/\\\"$NF }}'\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	m, err := modem.NewWithJsonBytes([]byte("{\"model\":\"rm500q\",\"findifacename\":\"ls -l /sys/class/net|grep '2-4:1.4'|awk -F'[/]' '{ print $NF }'\",\"findatdevpath\":\"ls -l /sys/class/tty/ttyUSB*|grep '2-4:1.3'|awk -F'[/]' '{ print \\\"/dev/\\\"$NF }'\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	// m, err := modem.NewWithJsonBytes([]byte("{\"model\":\"rm500q\",\"findifacename\":\"ls -l /sys/class/net|grep '2-3:1.4'|awk -F'[/]' '{ print $NF }'\",\"findatdevpath\":\"ls -l /sys/class/tty/ttyUSB*|grep '2-3:1.3'|awk -F'[/]' '{ print \\\"/dev/\\\"$NF }'\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	if err != nil {
		log.Error(err)
		return
	}
	mslice[0] = m
	// m, err = modem.NewWithJsonBytes([]byte("{\"model\":\"rm500q\",\"findifacename\":\"ohno\",\"findatdevpath\":\"haha\",\"quectel\":\"/usr/bin/quectel-CM\"}"))
	// if err != nil {
	// 	log.Error(err)
	// 	return
	// }
	// mslice[1] = m

	for i := 0; i < mtot; i++ {
		m = mslice[i]
		log.SetFormatter(m)
		if err = m.OpenWithLogger(log); err != nil {
			log.Error(err)
			return
		}
		defer m.Close()
		log.Debugf("%+v", m)

		wg.Add(1)
		go modem.Start(m, &wg)
	}

	time.Sleep(time.Second * 60)

	for i := 0; i < mtot; i++ {
		modem.Stop(mslice[i])
	}
	wg.Wait()
}
