package modem

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

type IModem interface {
	Format(entry *logrus.Entry) ([]byte, error)
	Open() error
	OpenWithLogger(logger *logrus.Logger) error
	Close() error
	ToJson() string
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
	MSTAT_HOTPLUGDETECT
	MSTAT_CHECK_SIMREADY
	MSTAT_CHECK_REGISTRATIONM
	MSTAT_CHECK_IP
	MSTAT_CHECK_GATEWAY
	MSTAT_LOOPING
	MSTAT_SOFTRESET
	MSTAT_HARDRESET
	MSTAT_LOOP_STOP
)

type Modem struct {
	CfgJsonBytes  []byte
	Model         string `json:"model"`
	FindIfaceName string `json:"findifacename"`
	FindATdevPath string `json:"findatdevpath"`
	Name          string `json:"name"`
	l             *logrus.Logger
	needStop      bool
	state         MState
	ifaceName     string
	atDevPath     string
	at            *serial.Port
	ips           []net.IP
	gw            net.IP
}

func (m *Modem) String() string {
	// return "CfgJsonBytes:'" + string(m.CfgJsonBytes) + "',\n Model:'" + m.Model + "',\n FindIfaceName:'" + m.FindATdevPath + "',\n FindIfaceName:'" + m.FindIfaceName + "'\n"
	return "Model:'" + m.Model + "',\n FindIfaceName:'" + m.FindIfaceName + "',\n FindIfaceName:'" + m.FindATdevPath + "'\n"
}
func (m *Modem) GoString() string {
	return m.String()
}
func (m *Modem) ToJson() string {
	var ipSS []string
	for _, ip := range m.ips {
		if ip.To4() != nil {
			ipSS = append(ipSS, "\""+ip.String()+"\"")
		}
	}
	var gwStr string
	if m.gw.To4() != nil {
		gwStr = m.gw.String()
	}
	return "{\"iface\":\"" + m.ifaceName + "\"," +
		"\"ips\":[" + strings.Join(ipSS, ",") + "]," +
		"\"gw\":\"" + gwStr +
		"\"}"
}

func (m *Modem) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer //设置buffer缓冲区
	if entry.Buffer == nil {
		b = &bytes.Buffer{}
	} else {
		b = entry.Buffer
	}
	//设置格式
	fmt.Fprintf(b, "%5s %s:%d %s %s %s\n", entry.Level, path.Base(entry.Caller.File), entry.Caller.Line, m.Model, m.Name, entry.Message)
	// path.Base(entry.Caller.Function), m.Model, m.name, entry.Message)
	return b.Bytes(), nil
}

func (m *Modem) Open() error {
	return m.OpenWithLogger(nil)
}
func (m *Modem) OpenWithLogger(logger *logrus.Logger) error {
	if logger != nil {
		m.l = logger
	} else {
		m.l = logrus.StandardLogger()
	}
	if err := json.Unmarshal(m.CfgJsonBytes, m); err != nil {
		m.l.Errorf("json.Unmarshal()->:%v", err)
		return err
	}
	return nil
}
func (m *Modem) Close() error {
	// m.l.Info("Close")
	return nil
}

func (m *Modem) Run(wg *sync.WaitGroup) error {
	m.l.Debug("run")
	for {
		if m.needStop {
			m.l.Info("needStop")
			break
		}
		time.Sleep(time.Second * 2)
		m.l.Info("runing")
	}
	wg.Done()
	m.l.Info("Done")
	return nil
}

// ls -l /sys/class/net |awk -F'[/]' '{if($9~/1-1:1.4/){ print $NF }}'
func (m *Modem) isIfaceNameChange() bool {
	cmd := exec.Command("bash", "-c", m.FindIfaceName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout // 标准输出
	cmd.Stderr = &stderr // 标准错误
	err := cmd.Run()
	outStr, errStr := strings.Replace(string(stdout.Bytes()), "\n", "", -1), strings.Replace(string(stderr.Bytes()), "\n", "", -1)
	m.l.Debugf("isIfaceNameChange cmd.Run(%+v)->%v,%s,%s", cmd, err, outStr, errStr)
	if strings.Compare(m.ifaceName, outStr) != 0 {
		m.l.Infof("iface:'%s'->'%s'", m.ifaceName, outStr)
		m.ifaceName = outStr
		return true
	}
	return false
}

// ls -l /sys/class/tty/ttyUSB*|awk -F'[/]' '{if($13~/1-1:1.3/){ print "/dev/"$NF }}' go exec不能用通配符
// ls -l /sys/class/tty/|awk -F'[/ ]' '{if($20~/1-1:1.3/){ print "/dev/"$NF }}'
func (m *Modem) isATdevPathChange() bool {
	cmd := exec.Command("bash", "-c", m.FindATdevPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout // 标准输出
	cmd.Stderr = &stderr // 标准错误
	err := cmd.Run()
	outStr, errStr := strings.Replace(string(stdout.Bytes()), "\n", "", -1), strings.Replace(string(stderr.Bytes()), "\n", "", -1)
	m.l.Debugf("isATdevPathChange cmd.Run(%+v)->%v,%s,%s", cmd, err, outStr, errStr)
	if strings.Compare(m.atDevPath, outStr) != 0 {
		m.l.Infof("atDevPath:'%s'->'%s'", m.atDevPath, outStr)
		m.atDevPath = outStr
		return true
	}
	return false
}
func (m *Modem) atClose() error {
	m.l.Debug("atClose")
	if m.at != nil {
		if err := m.at.Close(); err != nil {
			m.l.Errorf("m.at.Close()->%v", err)
			return err
		}
		m.at = nil
	}
	return nil
}
func (m *Modem) atOpen() error {
	c := &serial.Config{Name: m.atDevPath, Baud: 115200 /*, ReadTimeout: time.Second * 10*/}
	at, err := serial.OpenPort(c)
	if err != nil {
		m.l.Errorf("serial.OpenPort(%+v)->%+v", c, err)
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
		m.l.Errorf("m.at.Write(%q)->%d,%+v", wr, n, err)
		return n, err
	}
	time.Sleep(t)
	n, err = m.at.Read(rd)
	if err != nil {
		m.l.Errorf("m.at.Read()->%+v", err)
		return n, err
	}
	m.l.Debugf("%q->%q", wr, rd[:n])
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
		m.l.Info("atNoEcho()->ok")
		return nil
	}
	return errors.New("Unknow " + fmt.Sprintf("%q", buf[:n]))
}
func (m *Modem) atSoftReset() error {
	m.l.Debug("softreset")
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
func (m *Modem) hardReset() error {
	m.l.Debug("hardreset doNothing")
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
		m.l.Debug("atIsOK")
		return nil
	}
	return errors.New("Unknow " + fmt.Sprintf("%q", buf[:n]))
}

func (m *Modem) hotplugDetect() error {
	m.l.Debug("hotplugDetect")
	return nil
}

func (m *Modem) isSimReady() error {
	m.l.Debug("isSimReady")
	return nil
}

func (m *Modem) isRegistertion() error {
	m.l.Debug("isRegistertion")
	return nil
}

func (m *Modem) hasIP() error {
	var reterr error
	for i := 0; i < 10; i++ {
		iface, err := net.InterfaceByName(m.ifaceName)
		if err != nil {
			m.l.Error(err)
			reterr = err
			time.Sleep(time.Second)
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			m.l.Error(err)
			reterr = err
			time.Sleep(time.Second)
			continue
		}
		m.ips = nil
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				m.ips = append(m.ips, ipnet.IP)
				// m.l.Infof("append(%+v)->ok", ipnet.IP.String())
			}
		}
		if len(m.ips) > 0 {
			m.l.Infof("hasIP(%+v)->ok", m.ips)
			return nil
		}
		reterr = errors.New("no ip found")
		time.Sleep(time.Second)
	}
	return reterr
}

func (m *Modem) hasGateway() error {
	return errors.New("Function Not implementated")
}

func (m *Modem) isDialUp() error {
	m.l.Debug("isDialUp")
	destIPs := []string{"223.5.5.5", "8.8.8.8"}
	var err error
	for _, destIP := range destIPs {
		cmd := exec.Command("ping", destIP, "-I", m.ifaceName, "-c", "1", "-W", "3")
		err = cmd.Run() //只执行，不获取输出
		if err != nil { //ping失败
			m.l.Errorf("cmd.Run(%+v)->%v", cmd, err)
			continue
		}
		m.l.Infof("cmd.Run(%+v)->%v", cmd, err)
		return nil
	}
	m.ips, m.gw = nil, nil
	return err
}

func (m *Modem) loopStart() {
	m.l.Debug("Loop starting")
	m.needStop = false
}
func (m *Modem) loopStop() {
	m.l.Debug("Loop stoping")
	m.needStop = true
}
func Start(m IModem, wg *sync.WaitGroup) {
	m.loopStart()
	m.run(wg)
}
func Stop(m IModem) {
	m.loopStop()
}

func NewWithJsonBytes(jsonbytes []byte) (IModem, error) {
	raw := Modem{CfgJsonBytes: jsonbytes}
	if err := json.Unmarshal(jsonbytes, &raw); err != nil {
		fmt.Printf("json.Unmarshal()->:%v\n", err)
		return nil, err
	}
	switch raw.Model {
	case "ep06":
		return &M_qws_ep06{M_qws{Modem: raw}}, nil
	case "rm500q":
		return &M_qws_rm500q{M_qws{Modem: raw}}, nil
	default:
		panic("Unknow supported Model" + raw.Model)
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
