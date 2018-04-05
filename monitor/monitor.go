package metrics

type Metric interface {
	Register(name string, rcvr interface{}, metadata string) (err error)
	HandleConnAccept(name string, rcvr interface{}, metadata string) (err error)
}
