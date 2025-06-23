package scraper

import (
	"fmt"
	"regexp"
	"strings"
	"github.com/robertkrimen/otto"
)

var (
	// Regex to find the JS challenge logic.
	jsChallengeRegex = regexp.MustCompile(`setTimeout\(function\(\){\s+(var s,t,o,p,b,r,e,a,k,i,n,g,f.+?a\.value =.+?)\r?\n`)
	// Regex to extract the core expression from the challenge.
	jsExpressionRegex = regexp.MustCompile(`var s,t,o,p,b,r,e,a,k,i,n,g,f, .+?={"(.+?)":\+?(.+?)}`)
	// Regex to extract the second part of the challenge logic for evaluation.
	jsPassRegex = regexp.MustCompile(`a\.value = (.+ \+ t.length).toFixed\(10\)`)
)

// solveJSChallenge uses Otto to solve Cloudflare's JS math challenge.
func solveJSChallenge(body, domain string) (string, error) {
	// 1. Extract the main challenge script
	matches := jsChallengeRegex.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find Cloudflare JS challenge script")
	}
	challengeScript := matches[1]

	// 2. Extract the expression (e.g., "M": "+((!+[]...))")
	exprMatches := jsExpressionRegex.FindStringSubmatch(challengeScript)
	if len(exprMatches) < 3 {
		return "", fmt.Errorf("could not find JS challenge expression")
	}
	
	// This is the obfuscated JS math expression
	obfuscatedJS := exprMatches[2]

	// 3. Extract the final calculation logic
	passMatches := jsPassRegex.FindStringSubmatch(challengeScript)
	if len(passMatches) < 2 {
		return "", fmt.Errorf("could not find JS challenge pass expression")
	}
	finalCalc := passMatches[1]
	
	// Replace the variable name with its obfuscated value
	finalCalc = strings.Replace(finalCalc, "wKRocaN."+exprMatches[1], obfuscatedJS, 1)

	// 4. Execute the JS in Otto
	vm := otto.New()
	
	// The challenge script sometimes uses 't.length', which is the length of the domain.
	vm.Set("t", domain)
	
	// Evaluate the final calculation
	result, err := vm.Run(finalCalc)
	if err != nil {
		return "", fmt.Errorf("otto: failed to run JS: %w", err)
	}

	// 5. Format the answer to 10 decimal places, like the original script
	floatResult, err := result.ToFloat()
	if err != nil {
		return "", fmt.Errorf("otto: could not convert result to float: %w", err)
	}

	return fmt.Sprintf("%.10f", floatResult), nil
}