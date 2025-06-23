package js

import (
	_ "embed"
	"fmt"
	"github.com/Advik-B/cloudscraper/lib/errors"
	"log"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

// Create a Simulated Browser Environment (DOM Shim)
//
//go:embed setup.js
var setupScript string

// OttoEngine uses the embedded otto interpreter.
type OttoEngine struct{}

// NewOttoEngine creates a new engine that uses the built-in otto interpreter.
func NewOttoEngine() *OttoEngine {
	return &OttoEngine{}
}

// Run executes a script in otto. It captures output by overriding console.log.
func (e *OttoEngine) Run(script string) (string, error) {
	vm := otto.New()
	var result string

	// Setup safe console.log capturing
	err := vm.Set("console", map[string]interface{}{
		"log": func(call otto.FunctionCall) otto.Value {
			result = call.Argument(0).String()
			return otto.Value{}
		},
	})
	if err != nil {
		return "", fmt.Errorf("otto: failed to set console.log: %w", err)
	}

	// === Hardened Execution ===
	const maxExecutionTime = 3 * time.Second
	vm.Interrupt = make(chan func(), 1)
	watchdogDone := make(chan struct{})
	defer close(watchdogDone)

	// Watchdog goroutine to interrupt runaway code
	go func() {
		select {
		case <-time.After(maxExecutionTime):
			vm.Interrupt <- func() {
				panic(errors.ErrExecutionTimeout)
			}
		case <-watchdogDone:
		}
	}()

	// Recover from intentional interrupts
	defer func() {
		if r := recover(); r != nil {
			if r == errors.ErrExecutionTimeout {
				err = fmt.Errorf("otto: script execution timed out after %v", maxExecutionTime)
			} else {
				panic(r) // Bubble up unexpected panics
			}
		}
	}()

	_, err = vm.Run(script)
	if err != nil {
		return "", fmt.Errorf("otto: script execution failed: %w", err)
	}
	return result, nil
}

// SolveV2Challenge uses the original synchronous method to solve v2 challenges,
// as otto does not support asynchronous operations like setTimeout.
func (e *OttoEngine) SolveV2Challenge(body, domain string, scriptMatches [][]string, logger *log.Logger) (string, error) {
	vm := otto.New()

	// Security: Running setup script in VM.
	if _, err := vm.Run(setupScript); err != nil {
		return "", fmt.Errorf("otto: failed to set up DOM shim: %w", err)
	}

	// Execute all extracted Cloudflare scripts in the same VM context.
	for _, match := range scriptMatches {
		if len(match) > 1 {
			scriptContent := match[1]
			scriptContent = strings.ReplaceAll(scriptContent, `document.getElementById('challenge-form');`, "({})")
			// Security: This executes JavaScript from the Cloudflare challenge page.
			// The otto VM is sandboxed, but this is an inherent risk of the library's function.
			if _, err := vm.Run(scriptContent); err != nil {
				logger.Printf("otto: warning, a script block failed to run: %v\n", err)
			}
		}
	}

	// Wait for the script's internal timeouts to complete.
	time.Sleep(4 * time.Second)

	// Get the final answer from the 'jschl_answer' field in the dummy document.
	// Security: This executes a small, controlled script to retrieve a value.
	answerObj, err := vm.Run(`document.getElementById('jschl-answer').value`)
	if err != nil || !answerObj.IsString() {
		return "", fmt.Errorf("otto: could not retrieve final answer from VM: %w", err)
	}

	return answerObj.String(), nil
}
