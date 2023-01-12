package modem

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type M_qws struct {
	Modem
	BusType    string `json:"busType"`
	Quectel    string `json:"quectel"`
	cmd        *exec.Cmd
	checkCount uint8
	wg         sync.WaitGroup
}

func (m *M_qws) String() string {
	return fmt.Sprintf("Name:%q,Model:%q,BustType:%q,FindIfaceName:%q,FindATdevPath:%q", m.Name, m.Model, m.BusType, m.FindIfaceName, m.FindATdevPath)
}
func (m *M_qws) GoString() string {
	return m.String()
}

func (m *M_qws) Open() error {
	return m.OpenWithLogger(nil)
}
func (m *M_qws) OpenWithLogger(logger *logrus.Logger) error {
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
func (m *M_qws) run(wg *sync.WaitGroup) error {
	defer wg.Done()
OuterLop:
	for {
		delayTime := time.Millisecond * 50
		if m.needStop {
			m.l.Info("needStop")
			m.state = MSTAT_LOOP_STOP
		}

		switch m.state {
		case MSTAT_LOOP_STOP:
			m.atClose()
			m.stopQuectel()
			m.atDevPath = ""
			m.ifaceName, m.realIfaceName = "", ""
			break OuterLop
		case MSTAT_INIT, MSTAT_CHECK_IFACENAME_CHANGE:
			m.l.Debug("MSTAT_INIT,MSTAT_CHECK_IFACENAME_CHANGE")
			if err := m.findIfaceName(); err != nil {
				m.l.Error(err)
				delayTime = time.Second * 3
			} else {
				m.state = MSTAT_CHECK_ATDEVPATH_CHANGE
			}
		case MSTAT_CHECK_ATDEVPATH_CHANGE:
			m.l.Debug("MSTAT_CHECK_ATDEVPATH_CHANGE")
			if m.isATdevPathChange() {
				m.state = MSTAT_CLOSE_ATDEV
			} else {
				delayTime = time.Second * 3
			}
		case MSTAT_CLOSE_ATDEV:
			m.l.Debug("MSTAT_CLOSE_ATDEV")
			if err := m.atClose(); err != nil {
				m.l.Error(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_OPEN_ATDEV
			}
		case MSTAT_OPEN_ATDEV:
			m.l.Debug("MSTAT_OPEN_ATDEV")
			if err := m.atOpen(); err != nil {
				m.l.Error(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_NOECHO
			}
		case MSTAT_NOECHO:
			m.l.Debug("MSTAT_NOECHO")
			if err := m.atNoEcho(); err != nil {
				m.l.Error(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_HOTPLUGDETECT
			}
		case MSTAT_HOTPLUGDETECT:
			m.l.Debug("MSTAT_HOTPLUGDETECT")
			if err := m.hotplugDetect(); err != nil {
				m.l.Error(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_CHECK_SIMREADY
				m.checkCount = 0
			}
		case MSTAT_CHECK_SIMREADY:
			m.l.Debug("MSTAT_CHECK_SIMREADY")
			if err := m.isSimReady(); err != nil {
				if m.checkCount > 10 {
					m.checkCount = 0
					m.l.Error(err)
					m.state = MSTAT_SOFTRESET
				}
				m.checkCount++
				delayTime = time.Second * 3
			} else {
				m.checkCount = 0
				m.state = MSTAT_CHECK_REGISTRATIONM
			}
		case MSTAT_CHECK_REGISTRATIONM:
			m.l.Debug("MSTAT_CHECK_REGISTRATIONM")
			if err := m.isRegistertion(); err != nil {
				if m.checkCount > 10 {
					m.checkCount = 0
					m.l.Error(err)
					m.state = MSTAT_SOFTRESET
				}
				m.checkCount++
				delayTime = time.Second * 2
			} else {
				m.checkCount = 0
				m.state = MSTAT_QWS_STOP_QUEDTEL
			}
		case MSTAT_QWS_STOP_QUEDTEL:
			m.l.Debug("MSTAT_QWS_STOP_QUEDTEL")
			if err := m.stopQuectel(); err != nil {
				m.l.Error(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_QWS_START_QUEDTEL
			}
		case MSTAT_QWS_START_QUEDTEL:
			m.l.Debug("MSTAT_QWS_START_QUEDTEL")
			if err := m.startQuectel(); err != nil {
				m.l.Error(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_CHECK_IP
			}
		case MSTAT_CHECK_IP:
			m.l.Debug("MSTAT_CHECK_IP")
			if err := m.hasIP(); err != nil {
				if m.checkCount > 10 {
					m.checkCount = 0
					m.l.Error(err)
					m.state = MSTAT_SOFTRESET
				}
				m.checkCount++
				delayTime = time.Second * 2
			} else {
				m.checkCount = 0
				m.state = MSTAT_CHECK_GATEWAY
			}
		case MSTAT_CHECK_GATEWAY:
			m.l.Debug("MSTAT_CHECK_GATEWAY")
			if err := m.hasGateway(); err != nil {
				if m.checkCount > 10 {
					m.checkCount = 0
					m.l.Error(err)
					m.state = MSTAT_SOFTRESET
				}
				m.checkCount++
				delayTime = time.Second * 2
			} else {
				m.checkCount = 0
				m.state = MSTAT_LOOPING
			}

		case MSTAT_LOOPING:
			m.l.Debug("MSTAT_LOOPING")
			if err := m.atIsOK(); err != nil {
				m.checkCount = 0
				m.l.Error(err)
				m.state = MSTAT_SOFTRESET
				break
			}
			if err := m.isDialUp(); err != nil {
				m.l.Error(err)
				m.atClose()
				m.stopQuectel()
				m.atDevPath = ""
				m.ifaceName, m.realIfaceName = "", ""
				m.state = MSTAT_INIT
				break
			}
			delayTime = time.Second * 3

		case MSTAT_SOFTRESET:
			m.l.Debug("MSTAT_SOFTRESET")
			if err := m.atSoftReset(); err != nil {
				if m.checkCount > 5 {
					m.checkCount = 0
					m.l.Error(err)
					m.state = MSTAT_HARDRESET
				}
				m.checkCount++
				delayTime = time.Second * 3
			} else {
				m.checkCount = 0
				m.atClose()
				m.stopQuectel()
				m.atDevPath = ""
				m.ifaceName, m.realIfaceName = "", ""
				m.state = MSTAT_INIT
			}
		case MSTAT_HARDRESET:
			m.l.Debug("MSTAT_HARDRESET")
			if err := m.hardReset(); err != nil {
				m.l.Error(err)
			} else {
				m.state = MSTAT_INIT
			}
		}
		time.Sleep(delayTime)
		m.l.Debug("runing")
	}
	m.l.Info("Done")
	return nil
}

func (m *M_qws) findIfaceName() error {
	if err := m.Modem.findIfaceName(); err != nil {
		return err
	}
	if m.BusType == "pcie" {
		m.realIfaceName = m.ifaceName + ".1"
		m.l.Infof("realIfaceName:%q", m.realIfaceName)
	}
	return nil
}

func (m *M_qws) stopQuectel() error {
	if len(m.ifaceName) == 0 {
		return errors.New("no ifaceName found")
	}
	m.cmd = exec.Command("/usr/bin/pkill", "-f", m.Quectel+" -i "+m.ifaceName)
	err := m.cmd.Run()
	m.l.Infof("cmd.Run(%+v)->%v", m.cmd, err)
	m.wg.Wait()
	return nil
}
func (m *M_qws) startQuectel() error {
	if len(m.ifaceName) == 0 {
		return errors.New("no ifaceName found")
	}
	m.cmd = exec.Command("/usr/bin/pgrep", "-f", m.Quectel+" -i "+m.ifaceName)
	out, err := m.cmd.CombinedOutput()
	// if err != nil {
	// 	m.l.Debugf("cmd.Run(%+v)->%+v,%v", m.cmd, out, err)
	// }
	if err == nil && len(out) != 0 {
		m.l.Warnf("cmd.Run(%+v)->%+v,is already run", m.cmd, out)
		return nil
	}
	m.l.Infof("cmd.Run(%+v)->%+v,%v", m.cmd, out, err)

	m.cmd = exec.Command(m.Quectel, "-i", m.ifaceName, "&")
	go func() {
		m.wg.Add(1)
		defer m.wg.Done()
		m.l.Info("go quectel start")
		err := os.MkdirAll("/tmp/qws", os.ModePerm)
		if err != nil {
			logrus.Error(err)
		}
		stdout, err := os.OpenFile("/tmp/qws/"+m.ifaceName+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			m.l.Errorf("os.OpenFile(/tmp/qws/%s.log)->%v", m.ifaceName, err)
			return
		}
		defer stdout.Close()
		m.cmd.Stdout, m.cmd.Stderr = stdout, stdout
		err = m.cmd.Start()
		m.l.Infof("cmd.Start(%+v)->%+v", m.cmd, err)
		m.cmd.Wait()
		m.l.Info("go quectel stop")
	}()
	return nil
}

func (m *M_qws) hotplugDetect() error {
	atcmd := []byte("at+qsimdet=1,0\r\n")
	buf := make([]byte, 128)
	n, err := m.atWriteRead(atcmd, buf)
	if err != nil {
		return err
	}
	if bytes.Contains(buf, []byte("OK")) {
		m.l.Infof("ok")
		return nil
	}
	m.l.Warnf("Unknow [%q]", buf[:n])
	return errors.New("Unknow " + fmt.Sprintf("%q", buf[:n]))
}

func (m *M_qws) isSimReady() error {
	atcmd := []byte("at+cpin?\r\n")
	buf := make([]byte, 128)
	n, err := m.atWriteRead(atcmd, buf)
	if err != nil {
		return err
	}
	if bytes.Contains(buf, []byte("READY")) { //\r\n+CPIN: READY\r\nOK\r\n
		m.l.Infof("ok")
		return nil
	} else if bytes.Contains(buf, []byte("ERROR")) { //"\r\n+CME ERROR: 13\r\n" "\r\n+CME ERROR: 10\r\n"
		m.l.Warnf("ERROR,count%d", m.checkCount)
	} else {
		m.l.Errorf("Unknow %q", buf[:n])
		return errors.New("Unknow " + fmt.Sprintf("%q", buf[:n]))
	}
	return errors.New("SimIsNotReady")
}

func (m *M_qws) isRegistertion() error {
	return nil //TODO 暂未确定用哪个指令查
	// 	atcmd := []byte("at+cereg?\r\n")
	// 	buf := make([]byte, 128)
	// 	n, err := m.atWriteRead(atcmd, buf)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if bytes.Contains(buf, []byte("CEREG: 0,1")) || bytes.Contains(buf, []byte("CEREG: 0,5")) { //\r\n+CEREG: 0,1\r\nOK\r\n  或者0,5
	// 		m.l.Infof("ok")
	// 		return nil
	// 	}
	// 	m.l.Warnf("no %q,cnt%d", buf[:n], i)
	// 	time.Sleep(time.Second * 3)
	// return errors.New("isNotRegistertion")
}

func (m *M_qws) hasGateway() error {
	data, err := ioutil.ReadFile("/tmp/qws/" + m.realIfaceName + ".gw")
	if err != nil {
		m.gw = nil
		m.l.Error(err)
		return err
	}
	gwBS := bytes.Trim(data, "\n")
	if gw := net.ParseIP(string(gwBS)); gw.To4() != nil {
		m.gw = gw
		m.l.Infof("%+v ok", m.gw)
		return nil
	}
	return errors.New("no gateway found")
}

//TODO 读取信号
//static int m_qws_referencesignalreceivingpower(modem_t *m, char *e_out, uint16_t l4e_out)
//{
//    check_arg_null(e_out); check_arg_zero(l4e_out);
//    check_arg_null_o(e_out, l4e_out, m);
//    m_at_t *at = &m->at;
//    char rdbuf[1024] = {0}; //+QENG: "servingcell","NOCONN","LTE","FDD",460,11,760C614,268,2452,5,3,3,751B,-104,-11,-76,11,11\r\nOK
//    int rc = atmcallo(e_out, l4e_out, m_wr_rd, "at+qeng=\"servingcell\"\r\n", rdbuf, sizeof(rdbuf));
//    if(rc <= 0){ mle("rc:%d", rc); return rc; }
//    char *atarr[30] = {0};
//    int strcnt = atstr_slice(rdbuf, atarr, sizeof(atarr)/sizeof(char *));
//    // for(uint8_t i = 0; i < strcnt; i++){ mld("%q", atarr[i]); }
//    return E_OK;
//}
