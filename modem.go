package modem

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type IModem interface {
	String() string
	GoString() string
	Open() error
	Close() error
	// StartLoop() error
	// StopLoop() error
	// Run() error
	IsOK() error
	IsSimReady() error
	IsRegistertion() error
}

type Modem struct {
	CfgJsonBytes  []byte
	Model         string `json:"model"`
	FindIfaceName string `json:"findifacename"`
	FindATdevPath string `json:"findatdevpath"`
}

func (m *Modem) String() string {
	return "CfgJsonBytes:'" + string(m.CfgJsonBytes) + "',Model:'" + m.Model + "',FindIfaceName:'" + m.FindIfaceName + "',FindIfaceName:'" + m.FindIfaceName + "'"
}
func (m *Modem) GoString() string {
	return "CfgJsonBytes:'" + string(m.CfgJsonBytes) + "',Model:'" + m.Model + "',FindIfaceName:'" + m.FindIfaceName + "',FindIfaceName:'" + m.FindIfaceName + "'"
}
func (m *Modem) Open() error {
	fmt.Println("modem Open")
	if err := json.Unmarshal(m.CfgJsonBytes, m); err != nil {
		fmt.Printf("json.Unmarshal()->:%v", err)
		return err
	}
	return nil
}

func (m *Modem) Close() error {
	fmt.Println("modem Close")
	return nil
}

func (m *Modem) IsOK() error {
	fmt.Println("modem IsOK")
	return nil
}

func (m *Modem) IsSimReady() error {
	fmt.Println("modem IsSimReady")
	return nil
}

func (m *Modem) IsRegistertion() error {
	fmt.Println("modem IsRegistertion")
	return nil
}

func NewWithJsonBytes(jsonbytes []byte) (IModem, error) {
	fmt.Println(string(jsonbytes))
	rm := Modem{CfgJsonBytes: jsonbytes}
	if err := json.Unmarshal(jsonbytes, &rm); err != nil {
		fmt.Printf("json.Unmarshal()->:%v", err)
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
