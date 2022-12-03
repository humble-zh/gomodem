package modem

type Modem struct {
	Cfgjson string
}

func (m *Modem) IsOK() error {
	return nil
}

func (m *Modem) IsSimReady() error {
	return nil
}

func (m *Modem) IsRegistertion() error {
	return nil
}
