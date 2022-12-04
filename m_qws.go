package modem

import (
	"encoding/json"
	"fmt"
)

type M_qws struct {
	Modem
	Quectel string `json:"quectel"`
}

func (m *M_qws) Open() error {
	fmt.Println("qws Open")
	if err := json.Unmarshal(m.CfgJsonBytes, m); err != nil {
		fmt.Printf("json.Unmarshal()->:%v", err)
		return err
	}
	return nil
}

func (m *M_qws) Close() error {
	fmt.Println("qws Close")
	return nil
}

func (m *M_qws) IsOK() error {
	fmt.Println("qws IsOK")
	return nil
}

func (m *M_qws) IsSimReady() error {
	fmt.Println("qws IsSimReady")
	return nil
}

func (m *M_qws) IsRegistertion() error {
	fmt.Println("qws IsRegistertion")
	return nil
}

// func OpenFile() (Modem, error) {
// 	fmt.Println("hello")
// }
