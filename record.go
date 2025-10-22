package idempo

// Record holds the Action output.
type Record struct {
	ID            string
	InputHash     string
	SuccessOutput bool
	Output        []byte
}
