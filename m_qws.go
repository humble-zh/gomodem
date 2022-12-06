package modem

import (
	"encoding/json"
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
				m.state = MSTAT_LOOPING
			}
			//     rc = mmcall(m_noecho);
			//     if(E_OK == rc){ m->state = M_STATE_HOTPLUGDETECT; }
			//     else if(E_AT_RSP_ERROR == rc){ m->state = M_STATE_SOFTRESET; } // 模组回复ERROR
			//     else{ m->state = M_STATE_HARDRESET; } //模组回复超时或读写失败
			//     break;

			// case M_STATE_HOTPLUGDETECT:
			//     rc = mmcall(m_hotplugdetect);
			//     if(E_OK == rc){ m->state = M_STATE_CHECK_SIM_READY; }
			//     else if(E_AT_RSP_ERROR == rc){ m->state = M_STATE_SOFTRESET; } // 模组回复ERROR
			//     else{ m->state = M_STATE_HARDRESET; } //模组回复超时或读写失败
			//     break;
			// case M_STATE_CHECK_SIM_READY:
			//     rc = mmcall(m_simready);
			//     if(E_OK == rc){ m->state = M_STATE_CHECK_REGISTRATION; }
			//     else if(E_SIMABSENT == rc || E_AT_RSP_ERROR == rc){ m->state = M_STATE_SOFTRESET; }
			//     else{ m->state = M_STATE_HARDRESET; } //模组回复超时或读写失败
			//     break;
			// case M_STATE_CHECK_REGISTRATION:
			//     rc = mmcall(m_registration);
			//     if(E_OK == rc){ m->state = M_STATE_ROUTINE; }
			//     else if(E_UNREGISTRATION == rc){ m->state = M_STATE_SOFTRESET; }
			//     else{ m->state = M_STATE_HARDRESET; }
			//     break;

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
			if err := m.atHardReset(); err != nil {
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

func (m *M_qws) isSimReady() error {
	fmt.Printf("QWS %s isSimReady\n", m.Model)
	return nil
}

func (m *M_qws) isRegistertion() error {
	fmt.Printf("QWS %s isRegistertion\n", m.Model)
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
	fmt.Printf("QWS %s startQuectel\n", m.Model)
	m.cmd = exec.Command("/usr/bin/pgrep", "-f", m.Quectel+" -i "+m.ifacename+" -f /tmp/qws_"+m.ifacename+".log")
	fmt.Println(m.cmd.Args)
	out, err := m.cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd.Run(%+v)->%+v,%v\n", m.cmd, out, err)
	}
	fmt.Printf("cmd.Run(%+v)->%+v\n", m.cmd, out)
	if len(out) != 0 {
		return nil
	}

	m.cmd = exec.Command(m.Quectel, "-i", m.ifacename, "-f", "/tmp/qws_"+m.ifacename+".log", "&")
	fmt.Println(m.cmd.Args)
	go func() {
		_ = m.cmd.Start()
		m.cmd.Wait()
		fmt.Println("go stop")
	}()
	return nil
}
