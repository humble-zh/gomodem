package modem

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

type IModem interface {
	Open() error
	Close() error
	run(wg *sync.WaitGroup) error
	loopStart()
	loopStop()
}

type MState uint8

const (
	MSTAT_INIT MState = iota
	MSTAT_CHECK_IFACENAME_CHANGE
	MSTAT_QWS_STOP_QUEDTEL
	MSTAT_QWS_START_QUEDTEL
	MSTAT_CHECK_ATDEVPATH_CHANGE
	MSTAT_CLOSE_ATDEV
	MSTAT_OPEN_ATDEV

	MSTAT_NOECHO
	MSTAT_LOOPING
	MSTAT_SOFTRESET
	MSTAT_HARDRESET
	MSTAT_LOOP_STOP
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
	atdevpath string
	at        *serial.Port
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

// ls -l /sys/class/tty/ttyUSB*|awk -F'[/]' '{if($13~/1-1:1.3/){ print "/dev/"$NF }}' go exec不能用通配符
// ls -l /sys/class/tty/|awk -F'[/ ]' '{if($20~/1-1:1.3/){ print "/dev/"$NF }}'
func (m *Modem) isATdevPathChange() bool {
	fmt.Printf("Modem %s isATdevPathChange\n", m.Model)
	c1 := exec.Command("ls", "-l", "/sys/class/tty/")
	c2 := exec.Command("awk", "-F", "[/ ]", "{if($20~/"+m.USBbusPortID+".3/){ print \"/dev/\"$NF }}")
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
	if strings.Compare(m.atdevpath, outStr) != 0 {
		fmt.Printf("atdevpath:'%s'->'%s'\n", m.atdevpath, outStr)
		m.atdevpath = outStr
		return true
	}
	return false
}
func (m *Modem) atClose() error {
	fmt.Printf("Modem %s atClose\n", m.Model)
	if m.at != nil {
		if err := m.at.Close(); err != nil {
			fmt.Printf("Modem %s m.at.Close()->%v\n", m.Model, err)
			return err
		}
		m.at = nil
	}
	return nil
}
func (m *Modem) atOpen() error {
	c := &serial.Config{Name: m.atdevpath, Baud: 115200 /*, ReadTimeout: time.Second * 10*/}
	at, err := serial.OpenPort(c)
	if err != nil {
		fmt.Printf("serial.OpenPort(%+v)->%+v", c, err)
		return err
	}
	m.at = at
	return nil
}
func (m *Modem) atWriteReadTimeout(wr []byte, rd []byte, t time.Duration) (int, error) {
	if err := m.at.Flush(); err != nil {
		return -1, err
	}
	n, err := m.at.Write(wr)
	if err != nil {
		fmt.Printf("%s m.at.Write(%q)->%d,%+v\n", m.Model, wr, n, err)
		return n, err
	}
	time.Sleep(t)
	n, err = m.at.Read(rd)
	if err != nil {
		fmt.Printf("%s m.at.Read()->%+v\n", m.Model, err)
		return n, err
	}
	fmt.Printf("%s %q->%q\n", m.Model, wr, rd[:n])
	return n, nil
}
func (m *Modem) atWriteRead(wr []byte, rd []byte) (int, error) {
	return m.atWriteReadTimeout(wr, rd, time.Second*1)
}

func (m *Modem) atNoEcho() error {
	atcmd := []byte("ate0\r\n")
	buf := make([]byte, 128)
	n, err := m.atWriteRead(atcmd, buf)
	if err != nil {
		return err
	}
	if bytes.Contains(buf, []byte("OK")) {
		fmt.Printf("Modem %s atNoEcho()->ok\n", m.Model)
		return nil
	}
	return errors.New("Unknow " + fmt.Sprintf("%q", buf[:n]))
}
func (m *Modem) atSoftReset() error {
	fmt.Printf("Modem %s softreset\n", m.Model)
	atcmdcfun0 := []byte("at+cfun=0\r\n")
	bufcfun0 := make([]byte, 128)
	n, err := m.atWriteRead(atcmdcfun0, bufcfun0)
	if err != nil {
		return err
	}
	if !bytes.Contains(bufcfun0, []byte("OK")) {
		return errors.New("Unknow " + fmt.Sprintf("%q", bufcfun0[:n]))
	}

	atcmdcfun1 := []byte("at+cfun=1\r\n")
	bufcfun1 := make([]byte, 128)
	n, err = m.atWriteRead(atcmdcfun1, bufcfun1)
	if err != nil {
		return err
	}
	if bytes.Contains(bufcfun1, []byte("OK")) {
		return nil
	}
	return errors.New("Unknow " + fmt.Sprintf("%q", bufcfun1[:n]))
}
func (m *Modem) atHardReset() error {
	fmt.Printf("Modem %s hardreset doNothing\n", m.Model)
	return nil
}

func (m *Modem) atIsOK() error {
	atcmd := []byte("at\r\n")
	buf := make([]byte, 128)
	n, err := m.atWriteRead(atcmd, buf)
	if err != nil {
		return err
	}
	if bytes.Contains(buf, []byte("OK")) {
		fmt.Printf("Modem %s atIsOK\n", m.Model)
		return nil
	}
	return errors.New("Unknow " + fmt.Sprintf("%q", buf[:n]))
}

func (m *Modem) isSimReady() error {
	fmt.Printf("%s isSimReady\n", m.Model)
	return nil
}

func (m *Modem) isRegistertion() error {
	fmt.Printf("%s isRegistertion\n", m.Model)
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
