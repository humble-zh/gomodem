package modem

import (
	"fmt"
	"sync"
	"time"
)

type M_qws struct {
	Modem
	Quectel string `json:"quectel"`
}

func (m *M_qws) Run(wg *sync.WaitGroup) error {
	fmt.Printf("QWS %s run\n", m.Model)
	for {
		if m.needstop {
			fmt.Printf("QWS %s needstop\n", m.Model)
			break
		}
		time.Sleep(time.Second * 2)
		fmt.Printf("QWS %s runing\n", m.Model)
	}
	wg.Done()
	fmt.Printf("QWS %s Done\n", m.Model)
	return nil
}

func (m *M_qws) IsOK() error {
	fmt.Printf("QWS %s IsOK\n", m.Model)
	return nil
}

func (m *M_qws) IsSimReady() error {
	fmt.Printf("QWS %s IsSimReady\n", m.Model)
	return nil
}

func (m *M_qws) IsRegistertion() error {
	fmt.Printf("QWS %s IsRegistertion\n", m.Model)
	return nil
}
