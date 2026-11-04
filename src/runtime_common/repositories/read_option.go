package repositories

// ReadOption tunes how a repository read runs.
type ReadOption int

const (
	// ToLock holds the rows the read returns until the unit of work ends; outside one it locks nothing.
	ToLock ReadOption = iota + 1
)
