package js

import (
	"fmt"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

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
	err := vm.Set("console", map[string]interface{}{
		"log": func(call otto.FunctionCall) otto.Value {
			result = call.Argument(0).String()
			return otto.Value{}
		},
	})
	if err != nil {
		return "", fmt.Errorf("otto: failed to set console.log: %w", err)
	}

	_, err = vm.Run(script)
	if err != nil {
		return "", fmt.Errorf("otto: script execution failed: %w", err)
	}
	return result, nil
}

// SolveV2Challenge uses the original synchronous method to solve v2 challenges,
// as otto does not support asynchronous operations like setTimeout.
func (e *OttoEngine) SolveV2Challenge(body, domain string, scriptMatches [][]string) (string, error) {
	vm := otto.New()

	// Create a Simulated Browser Environment (DOM Shim)
	setupScript := `
		var window = this;
		var navigator = { userAgent: "" };
		var document = {
			getElementById: function(id) {
				return { value: "" };
			},
			createElement: function(tag) {
				return {
					firstChild: { href: "https://` + domain + `/" }
				};
			},
			cookie: ""
		};
		var atob = function(str) {
			var chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
			var a, b, c, d, e, f, g, i = 0, result = '';
			str = str.replace(/[^A-Za-z0-9\+\/\=]/g, '');
			do {
				a = chars.indexOf(str.charAt(i++)); b = chars.indexOf(str.charAt(i++)); c = chars.indexOf(str.charAt(i++)); d = chars.indexOf(str.charAt(i++));
				e = a << 18 | b << 12 | c << 6 | d; f = e >> 16 & 255; g = e >> 8 & 255; a = e & 255;
				result += String.fromCharCode(f);
				if (c != 64) result += String.fromCharCode(g);
				if (d != 64) result += String.fromCharCode(a);
			} while (i < str.length);
			return result;
		};
	`
	if _, err := vm.Run(setupScript); err != nil {
		return "", fmt.Errorf("otto: failed to set up DOM shim: %w", err)
	}

	// Execute all extracted Cloudflare scripts in the same VM context.
	for _, match := range scriptMatches {
		if len(match) > 1 {
			scriptContent := match[1]
			scriptContent = strings.ReplaceAll(scriptContent, `document.getElementById('challenge-form');`, "({})")
			if _, err := vm.Run(scriptContent); err != nil {
				fmt.Printf("otto: warning, a script block failed to run: %v\n", err)
			}
		}
	}

	// Wait for the script's internal timeouts to complete.
	time.Sleep(4 * time.Second)

	// Get the final answer from the 'jschl_answer' field in the dummy document.
	answerObj, err := vm.Run(`document.getElementById('jschl-answer').value`)
	if err != nil || !answerObj.IsString() {
		return "", fmt.Errorf("otto: could not retrieve final answer from VM: %w", err)
	}

	return answerObj.String(), nil
}
