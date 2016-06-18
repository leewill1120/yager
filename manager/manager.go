package manager

type Manager struct {
	ListenPort   int
	RegisterCode string
}

func NewManager(listenport int, registercode string) *Manager {
	m := &Manager{
		ListenPort:   listenport,
		RegisterCode: registercode,
	}
	return m
}

func (m *Manager) Run() {

}
