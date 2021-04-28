package applicator

type RaftApplicator interface {
	Apply(state []byte) error
}
