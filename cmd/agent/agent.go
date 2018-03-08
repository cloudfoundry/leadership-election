package agent

type Agent struct {
	IsLeader bool
}

func NewAgent() *Agent {
	return &Agent{
		IsLeader: true,
	}
}
