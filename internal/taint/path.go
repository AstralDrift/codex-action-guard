package taint

type Path struct {
	Source Source
	Sink   Sink
	Notes  []string
}
