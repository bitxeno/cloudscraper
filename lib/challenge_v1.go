package cloudscraper

import (
	"fmt"
	"github.com/Advik-B/cloudscraper/lib/errors"
	"github.com/robertkrimen/otto"
	"regexp"
	"strings"
)

var (
	jsV1ChallengeRegex  = regexp.MustCompile(`setTimeout\(function\(\){\s+(var s,t,o,p,b,r,e,a,k,i,n,g,f.+?a\.value =.+?)\r?\n`)
	jsV1ExpressionRegex = regexp.MustCompile(`var s,t,o,p,b,r,e,a,k,i,n,g,f, .+?={"(.+?)":\+?(.+?)}`)
	jsV1PassRegex       = regexp.MustCompile(`a\.value = (.+?)\.toFixed\(10\)`)
)

func solveV1Challenge(body, domain string) (string, error) {
	matches := jsV1ChallengeRegex.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find Cloudflare lib JS challenge script: %w", errors.ErrChallenge)
	}
	challengeScript := matches[1]

	exprMatches := jsV1ExpressionRegex.FindStringSubmatch(challengeScript)
	if len(exprMatches) < 3 {
		return "", fmt.Errorf("could not find JS lib challenge expression: %w", errors.ErrChallenge)
	}

	obfuscatedJS := exprMatches[2]
	finalCalcMatches := jsV1PassRegex.FindStringSubmatch(challengeScript)
	if len(finalCalcMatches) < 2 {
		return "", fmt.Errorf("could not find JS lib challenge pass expression: %w", errors.ErrChallenge)
	}

	finalCalc := finalCalcMatches[1]
	finalCalc = strings.Replace(finalCalc, exprMatches[0][strings.LastIndex(exprMatches[0], "{")+1:strings.Index(exprMatches[0], `":`)], obfuscatedJS, 1)

	vm := otto.New()
	vm.Set("t", domain)

	result, err := vm.Run(finalCalc)
	if err != nil {
		return "", fmt.Errorf("otto: failed to run lib JS: %w", err)
	}

	floatResult, err := result.ToFloat()
	if err != nil {
		return "", fmt.Errorf("otto: could not convert lib result to float: %w", err)
	}

	return fmt.Sprintf("%.10f", floatResult), nil
}
