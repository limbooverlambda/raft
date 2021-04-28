package member

type Entry struct {
	ID string
}

type RaftMember interface {
	List() ([]Entry, error)
}