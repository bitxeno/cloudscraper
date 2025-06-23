package js

// Engine defines the interface for a JavaScript runtime.
type Engine interface {
	// Run executes a self-contained JavaScript script and returns the result from stdout.
	Run(script string) (string, error)
}
