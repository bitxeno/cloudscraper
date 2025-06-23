package scraper

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

// Regex to find and extract the modern challenge script content and data.
var v2ScriptRegex = regexp.MustCompile(`(?s)<script[^>]*>(.*?window\._cf_chl_opt.*?)<\/script>`)

// solveModernChallenge uses a pure Go approach with otto to solve v2/v3 challenges.
func solveModernChallenge(body, domain string) (string, error) {
	// Find all script blocks, as Cloudflare may split logic.
	matches := v2ScriptRegex.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return "", fmt.Errorf("could not find modern JS challenge scripts")
	}

	vm := otto.New()

	// --- Create a Simulated Browser Environment (DOM Shim) ---
	setupScript := `
		var window = this;
		var navigator = { userAgent: "" }; // Will be set from scraper options
		var document = {
			getElementById: function(id) {
				// Return a dummy object with a 'value' property so expressions like 'y.value = ...' don't fail.
				return { value: "" };
			},
			createElement: function(tag) {
				return {
					// Cloudflare checks the href of a created 'a' tag.
					// We must provide it with a valid-looking URL.
					firstChild: { href: "https://` + domain + `/" }
				};
			},
			// Dummy cookie property for the script to write to.
			cookie: ""
		};
		// Polyfill for atob, which is used by Cloudflare but not native to otto.
		var atob = function(str) {
			// This is a basic base64 decoder polyfill for JS environments.
			var chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
			var a, b, c, d, e, f, g, i = 0, result = '';
			str = str.replace(/[^A-Za-z0-9\+\/\=]/g, '');
			do {
				a = chars.indexOf(str.charAt(i++));
				b = chars.indexOf(str.charAt(i++));
				c = chars.indexOf(str.charAt(i++));
				d = chars.indexOf(str.charAt(i++));
				e = a << 18 | b << 12 | c << 6 | d;
				f = e >> 16 & 255;
				g = e >> 8 & 255;
				a = e & 255;
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
	// --------------------------------------------------------

	// Execute all extracted Cloudflare scripts in the same VM context.
	for _, match := range matches {
		if len(match) > 1 {
			scriptContent := match[1]
			// The script expects a 'challenge-form' to exist for submission. We stub it.
			scriptContent = strings.ReplaceAll(scriptContent, `document.getElementById('challenge-form');`, "({})")
			if _, err := vm.Run(scriptContent); err != nil {
				// This might fail on some scripts that are part of the chain, which can be okay.
				// We only care about the final answer.
				fmt.Printf("otto: warning, a script block failed to run: %v\n", err)
			}
		}
	}

	// The answer is often stored in 'window._cf_chl_opt.cRq' or a similar structure.
	// We wait briefly for any setTimeout calls to complete.
	time.Sleep(4 * time.Second)

	// Try to get the final answer from the 'jschl_answer' field in the dummy document.
	answerObj, err := vm.Run(`document.getElementById('jschl-answer').value`)
	if err != nil || !answerObj.IsString() {
		return "", fmt.Errorf("otto: could not retrieve final answer from VM: %w", err)
	}

	return answerObj.String(), nil
}