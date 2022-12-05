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
	for {
		if m.needstop {
			fmt.Printf("QWS %s needstop\n", m.Model)
			break
		}

		switch m.state {
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
			}
			m.state = MSTAT_QWS_START_QUEDTEL
		case MSTAT_QWS_START_QUEDTEL:
			fmt.Println("MSTAT_QWS_START_QUEDTEL")
			if err := m.startQuectel(); err != nil {
				fmt.Println(err)
				m.state = MSTAT_INIT
			}
			m.state = MSTAT_LOOPING
		case MSTAT_LOOPING:
			fmt.Println("MSTAT_LOOPING")
		}

		time.Sleep(time.Second * 2)
		fmt.Printf("QWS %s runing\n", m.Model)
	}
	wg.Done()
	fmt.Printf("QWS %s Done\n", m.Model)
	return nil
}

func (m *M_qws) isOK() error {
	fmt.Printf("QWS %s isOK\n", m.Model)
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
