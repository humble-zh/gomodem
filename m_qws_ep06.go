package modem

type m_qws_ep06 struct {
	m_qws
}

func (m *m_qws_ep06) IsOK() error {
	return nil
}

func (m *m_qws_ep06) IsSimReady() error {
	return nil
}

func (m *m_qws_ep06) IsRegistertion() error {
	return nil
}
