package main

import (
	"fmt"
	"github.com/Advik-B/cloudscraper/errors"
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
		return "", fmt.Errorf("could not find Cloudflare v1 JS challenge script: %w", errors.ErrChallenge)
	}
	challengeScript := matches[1]

	exprMatches := jsV1ExpressionRegex.FindStringSubmatch(challengeScript)
	if len(exprMatches) < 3 {
		return "", fmt.Errorf("could not find JS v1 challenge expression: %w", errors.ErrChallenge)
	}

	obfuscatedJS := exprMatches[2]
	finalCalcMatches := jsV1PassRegex.FindStringSubmatch(challengeScript)
	if len(finalCalcMatches) < 2 {
		return "", fmt.Errorf("could not find JS v1 challenge pass expression: %w", errors.ErrChallenge)
	}

	finalCalc := finalCalcMatches[1]
	finalCalc = strings.Replace(finalCalc, exprMatches[0][strings.LastIndex(exprMatches[0], "{")+1:strings.Index(exprMatches[0], `":`)], obfuscatedJS, 1)

	vm := otto.New()
	vm.Set("t", domain)

	result, err := vm.Run(finalCalc)
	if err != nil {
		return "", fmt.Errorf("otto: failed to run v1 JS: %w", err)
	}

	floatResult, err := result.ToFloat()
	if err != nil {
		return "", fmt.Errorf("otto: could not convert v1 result to float: %w", err)
	}

	return fmt.Sprintf("%.10f", floatResult), nil
}
