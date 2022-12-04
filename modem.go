package modem

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
)

type IModem interface {
	String() string
	GoString() string
	Open() error
	Close() error
	Run(wg *sync.WaitGroup) error
	IsOK() error
	IsSimReady() error
	IsRegistertion() error
	loopStart()
	loopStop()
}

type Modem struct {
	CfgJsonBytes  []byte
	Model         string `json:"model"`
	FindIfaceName string `json:"findifacename"`
	FindATdevPath string `json:"findatdevpath"`
	needstop      bool
}

func (m *Modem) String() string {
	return "CfgJsonBytes:'" + string(m.CfgJsonBytes) + "',Model:'" + m.Model + "',FindIfaceName:'" + m.FindIfaceName + "',FindIfaceName:'" + m.FindIfaceName + "'"
}
func (m *Modem) GoString() string {
	return "CfgJsonBytes:'" + string(m.CfgJsonBytes) + "',Model:'" + m.Model + "',FindIfaceName:'" + m.FindIfaceName + "',FindIfaceName:'" + m.FindIfaceName + "'"
}
func (m *Modem) Open() error {
	// fmt.Println("modem Open")
	if err := json.Unmarshal(m.CfgJsonBytes, m); err != nil {
		fmt.Printf("json.Unmarshal()->:%v\n", err)
		return err
	}
	return nil
}
func (m *Modem) Close() error {
	// fmt.Printf("%s Close\n", m.Model)
	return nil
}

func (m *Modem) Run(wg *sync.WaitGroup) error {
	fmt.Printf("%s run\n", m.Model)
	for {
		if m.needstop {
			fmt.Printf("%s needstop\n", m.Model)
			break
		}
		time.Sleep(time.Second * 2)
		fmt.Printf("%s runing\n", m.Model)
	}
	wg.Done()
	fmt.Printf("%s Done\n", m.Model)
	return nil
}

func (m *Modem) IsOK() error {
	fmt.Printf("%s IsOK\n", m.Model)
	return nil
}

func (m *Modem) IsSimReady() error {
	fmt.Printf("%s IsSimReady\n", m.Model)
	return nil
}

func (m *Modem) IsRegistertion() error {
	fmt.Printf("%s IsRegistertion\n", m.Model)
	return nil
}

func (m *Modem) loopStart() {
	fmt.Printf("%s Loop starting\n", m.Model)
	m.needstop = false
}
func (m *Modem) loopStop() {
	fmt.Printf("%s Loop stoping\n", m.Model)
	m.needstop = true
}
func Start(m IModem, wg *sync.WaitGroup) {
	m.loopStart()
	m.Run(wg)
}
func Stop(m IModem) {
	m.loopStop()
}

func NewWithJsonBytes(jsonbytes []byte) (IModem, error) {
	// fmt.Println(string(jsonbytes))
	rm := Modem{CfgJsonBytes: jsonbytes}
	if err := json.Unmarshal(jsonbytes, &rm); err != nil {
		fmt.Printf("json.Unmarshal()->:%v\n", err)
		return nil, err
	}
	switch rm.Model {
	case "ep06":
		return &M_qws_ep06{M_qws{rm, ""}}, nil
	case "rm500q":
		return &M_qws_rm500q{M_qws{rm, ""}}, nil
	default:
		panic("Unknow supported Model" + rm.Model)
	}
}

func NewWithJsonFile(jsonfile string) (IModem, error) {
	fmt.Println(jsonfile)
	jsonbytes, err := ioutil.ReadFile(jsonfile)
	if err != nil {
		fmt.Printf("ioutil.ReadFile()->:%v\n", err)
		return nil, err
	}
	return NewWithJsonBytes(jsonbytes)
}
