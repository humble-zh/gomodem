package modem

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type IModem interface {
	Open() error
	Close() error
	run(wg *sync.WaitGroup) error
	isOK() error
	isSimReady() error
	isRegistertion() error
	loopStart()
	loopStop()
}

type MState uint8

const (
	MSTAT_INIT MState = iota
	MSTAT_CHECK_IFACENAME_CHANGE
	MSTAT_QWS_STOP_QUEDTEL
	MSTAT_QWS_START_QUEDTEL
	MSTAT_LOOPING
)

type Modem struct {
	CfgJsonBytes []byte
	Model        string `json:"model"`
	USBbusPortID string `json:"usbbusportid"`
	// FindIfaceName string `json:"findifacename"`
	// FindATdevPath string `json:"findatdevpath"`
	needstop  bool
	state     MState
	ifacename string
}

func (m *Modem) String() string {
	// return "CfgJsonBytes:'" + string(m.CfgJsonBytes) + "',\n Model:'" + m.Model + "',\n FindIfaceName:'" + m.FindATdevPath + "',\n FindIfaceName:'" + m.FindIfaceName + "'\n"
	// return "Model:'" + m.Model + "',\n FindIfaceName:'" + m.FindIfaceName + "',\n FindIfaceName:'" + m.FindATdevPath + "'\n"
	return "Model:'" + m.Model + "',\n USBbusPortID:'" + m.USBbusPortID + "'\n"
}
func (m *Modem) GoString() string {
	return m.String()
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

// ls -l /sys/class/net |awk -F'[/]' '{if($9~/1-1:1.4/){ print $NF }}'
func (m *Modem) isIfaceNameChange() bool {
	fmt.Printf("Modem %s isIfaceNameChange\n", m.Model)
	c1 := exec.Command("ls", "-l", "/sys/class/net")
	c2 := exec.Command("awk", "-F", "[/]", "{if($9~/"+m.USBbusPortID+".4/){ print $NF }}")
	c2.Stdin, _ = c1.StdoutPipe() //把c1的输出作为c2的输入
	var stdout, stderr bytes.Buffer
	c2.Stdout = &stdout // 标准输出
	c2.Stderr = &stderr // 标准错误
	fmt.Printf("%+v|%+v\n", c1, c2)
	if err := c2.Start(); err != nil {
		fmt.Printf("%+v\n", err)
	}
	if err := c1.Run(); err != nil {
		fmt.Printf("%+v\n", err)
	}
	if err := c2.Wait(); err != nil {
		fmt.Printf("%+v\n", err)
	}

	outStr, errStr := strings.Replace(string(stdout.Bytes()), "\n", "", -1), strings.Replace(string(stderr.Bytes()), "\n", "", -1)
	fmt.Printf("out:'%s',err:'%s'\n", outStr, errStr)
	if strings.Compare(m.ifacename, outStr) != 0 {
		fmt.Printf("iface:'%s'->'%s'\n", m.ifacename, outStr)
		m.ifacename = outStr
		return true
	}
	return false
}

//ls -l /sys/class/tty/ttyUSB*|awk -F'[/]' '{if($13~/1-1:1.3/){ print "/dev/"$NF }}'

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
	m.run(wg)
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
		return &M_qws_ep06{M_qws{rm, "", nil}}, nil
	case "rm500q":
		return &M_qws_rm500q{M_qws{rm, "", nil}}, nil
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
