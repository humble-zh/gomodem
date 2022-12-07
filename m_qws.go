package modem

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

type M_qws struct {
	Modem
	Quectel string `json:"quectel"`
	cmd     *exec.Cmd
}

func (m *M_qws) String() string {
	// return "CfgJsonBytes:'" + string(m.CfgJsonBytes) + "',\n Model:'" + m.Model + "',\n FindIfaceName:'" + m.FindATdevPath + "',\n FindIfaceName:'" + m.FindIfaceName + "'\n"
	// return "Model:'" + m.Model + "',\n FindIfaceName:'" + m.FindIfaceName + "',\n FindIfaceName:'" + m.FindATdevPath + "',\n Quectel:'" + m.Quectel + "'\n"
	return "Model:'" + m.Model + "',\n USBbusPortID:'" + m.USBbusPortID + "',\n Quectel:'" + m.Quectel + "'\n"
}
func (m *M_qws) GoString() string {
	return m.String()
}

func (m *M_qws) Open() error {
	// fmt.Println("M_qws Open")
	if err := json.Unmarshal(m.CfgJsonBytes, m); err != nil {
		fmt.Printf("json.Unmarshal()->:%v\n", err)
		return err
	}
	return nil
}
func (m *M_qws) run(wg *sync.WaitGroup) error {
	fmt.Printf("QWS %s run\n", m.Model)
OuterLop:
	for {
		if m.needstop {
			fmt.Printf("QWS %s needstop\n", m.Model)
			m.state = MSTAT_LOOP_STOP
		}

		switch m.state {
		case MSTAT_LOOP_STOP:
			m.atdevpath = ""
			m.ifacename = ""
			m.atClose()
			m.stopQuectel()
			break OuterLop
		case MSTAT_INIT, MSTAT_CHECK_IFACENAME_CHANGE:
			fmt.Println("MSTAT_INIT,MSTAT_CHECK_IFACENAME_CHANGE")
			if m.isIfaceNameChange() {
				m.state = MSTAT_QWS_STOP_QUEDTEL
			}
		case MSTAT_QWS_STOP_QUEDTEL:
			fmt.Println("MSTAT_QWS_STOP_QUEDTEL")
			if err := m.stopQuectel(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_QWS_START_QUEDTEL
			}
		case MSTAT_QWS_START_QUEDTEL:
			fmt.Println("MSTAT_QWS_START_QUEDTEL")
			if err := m.startQuectel(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_CHECK_ATDEVPATH_CHANGE
			}

		case MSTAT_CHECK_ATDEVPATH_CHANGE:
			fmt.Println("MSTAT_CHECK_ATDEVPATH_CHANGE")
			if m.isATdevPathChange() {
				m.state = MSTAT_CLOSE_ATDEV
			}
		case MSTAT_CLOSE_ATDEV:
			fmt.Println("MSTAT_CLOSE_ATDEV")
			if err := m.atClose(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_OPEN_ATDEV
			}
		case MSTAT_OPEN_ATDEV:
			fmt.Println("MSTAT_OPEN_ATDEV")
			if err := m.atOpen(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_NOECHO
			}

		case MSTAT_NOECHO:
			fmt.Println("MSTAT_NOECHO")
			if err := m.atNoEcho(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_HOTPLUGDETECT
			}
		case MSTAT_HOTPLUGDETECT:
			fmt.Println("MSTAT_HOTPLUGDETECT")
			if err := m.hotplugDetect(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_INIT
			} else {
				m.state = MSTAT_CHECK_SIMREADY
			}
		case MSTAT_CHECK_SIMREADY:
			fmt.Println("MSTAT_CHECK_SIMREADY")
			if err := m.isSimReady(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_SOFTRESET
			} else {
				m.state = MSTAT_CHECK_REGISTRATIONM
			}
		case MSTAT_CHECK_REGISTRATIONM:
			fmt.Println("MSTAT_CHECK_REGISTRATIONM")
			if err := m.isRegistertion(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_SOFTRESET
			} else {
				m.state = MSTAT_LOOPING
			}

		case MSTAT_LOOPING:
			fmt.Println("MSTAT_LOOPING")
			if err := m.atIsOK(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_SOFTRESET
			} else {
				m.state = MSTAT_LOOPING
			}

		case MSTAT_SOFTRESET:
			fmt.Println("MSTAT_SOFTRESET")
			if err := m.atSoftReset(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_HARDRESET
			} else {
				m.atdevpath = ""
				m.ifacename = ""
				m.atClose()
				m.stopQuectel()
				m.state = MSTAT_INIT
			}
		case MSTAT_HARDRESET:
			fmt.Println("MSTAT_HARDRESET")
			if err := m.hardReset(); err != nil {
				fmt.Println(err)
			} else {
				m.state = MSTAT_INIT
			}
		}

		time.Sleep(time.Second * 2)
		fmt.Printf("QWS %s runing\n", m.Model)
	}
	wg.Done()
	fmt.Printf("QWS %s Done\n", m.Model)
	return nil
}

func (m *M_qws) stopQuectel() error {
	fmt.Printf("QWS %s stopQuectel\n", m.Model)
	m.cmd = exec.Command("/usr/bin/pkill", "-f", m.Quectel+" -i "+m.ifacename+" -f /tmp/qws_"+m.ifacename+".log")
	fmt.Println(m.cmd.Args)
	m.cmd.Run()
	return nil
}
func (m *M_qws) startQuectel() error {
	m.cmd = exec.Command("/usr/bin/pgrep", "-f", m.Quectel+" -i "+m.ifacename+" -f /tmp/qws_"+m.ifacename+".log")
	out, err := m.cmd.CombinedOutput()
	// if err != nil {
	// 	fmt.Printf("QWS %s startQuectel() cmd.Run(%+v)->%+v,%v\n", m.Model, m.cmd, out, err)
	// }
	if err == nil && len(out) != 0 {
		fmt.Printf("QWS %s startQuectel() cmd.Run(%+v)->%+v,is already run\n", m.Model, m.cmd, out)
		return nil
	}
	fmt.Printf("QWS %s startQuectel() cmd.Run(%+v)->%+v,%v\n", m.Model, m.cmd, out, err)

	m.cmd = exec.Command(m.Quectel, "-i", m.ifacename, "-f", "/tmp/qws_"+m.ifacename+".log", "&")
	go func() {
		err = m.cmd.Start()
		fmt.Printf("QWS %s startQuectel() cmd.Start(%+v)->%+v\n", m.Model, m.cmd, err)
		m.cmd.Wait()
		fmt.Println("go stop")
	}()
	return nil
}

func (m *M_qws) hotplugDetect() error {
	atcmd := []byte("at+qsimdet=1,1\r\n")
	buf := make([]byte, 128)
	n, err := m.atWriteRead(atcmd, buf)
	if err != nil {
		return err
	}
	if bytes.Contains(buf, []byte("OK")) {
		fmt.Printf("QWS %s hotplugDetect()->ok\n", m.Model)
		return nil
	}
	fmt.Printf("QWS %s hotplugDetect()->Unknow [%q]\n", m.Model, buf[:n])
	return errors.New("Unknow " + fmt.Sprintf("%q", buf[:n]))
}

func (m *M_qws) isSimReady() error {
	for i := 0; i < 10; i++ {
		atcmd := []byte("at+cpin?\r\n")
		buf := make([]byte, 128)
		n, err := m.atWriteRead(atcmd, buf)
		if err != nil {
			return err
		}
		if bytes.Contains(buf, []byte("READY")) { //\r\n+CPIN: READY\r\nOK\r\n
			fmt.Printf("QWS %s isSimReady()->ok\n", m.Model)
			return nil
		} else if bytes.Contains(buf, []byte("ERROR")) { //"\r\n+CME ERROR: 13\r\n" "\r\n+CME ERROR: 10\r\n"
			fmt.Printf("QWS %s isSimReady()->ERROR,cnt%d\n", m.Model, i)
			time.Sleep(time.Second * 3)
		} else {
			fmt.Printf("QWS %s isSimReady()->Unknow [%q]\n", m.Model, buf[:n])
			return errors.New("Unknow " + fmt.Sprintf("%q", buf[:n]))
		}
	}
	return errors.New("SimIsNotReady")
}

func (m *M_qws) isRegistertion() error {
	for i := 0; i < 10; i++ {
		atcmd := []byte("at+cereg?\r\n")
		buf := make([]byte, 128)
		n, err := m.atWriteRead(atcmd, buf)
		if err != nil {
			return err
		}
		if bytes.Contains(buf, []byte("CEREG: 0,1")) || bytes.Contains(buf, []byte("CEREG: 0,5")) { //\r\n+CEREG: 0,1\r\nOK\r\n  或者0,5
			fmt.Printf("QWS %s isRegistertion()->ok\n", m.Model)
			return nil
		}
		fmt.Printf("QWS %s isRegistertion()->no [%q],cnt%d\n", m.Model, buf[:n], i)
		time.Sleep(time.Second * 3)
	}
	return errors.New("isNotRegistertion")
}

//TODO 读取信号
