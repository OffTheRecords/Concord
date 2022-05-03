package main

import (
	"Concord/Authentication"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {

	//Get runtime args
	//Get log level
	logLevel, err := readStringArg("logging", "debug|error|warning|info")
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Printf("read argument %s as: %s\n", "logging", logLevel)

	fmt.Printf("%s\n", Authentication.Login())

	startRestAPI()
}

//Reads an argument and returns it if it matches the regex
func readStringArg(argName string, argRegexStr string) (string, error) {
	args := os.Args[1:] //Skip first arg of program name

	for i := 0; i < len(args); i++ {
		if strings.ToLower(args[i]) == "-"+strings.ToLower(argName) {
			if i+1 < len(args) {
				value := args[i+1]
				if !strings.Contains(value, "-") {
					argMatched, err := regexp.MatchString(argRegexStr, value)
					if err != nil {
						return "", err
					}
					if argMatched {
						return value, nil
					} else {
						return "", errors.New("error reading arguments, argument " + args[i+1] + " did not match the regex pattern " + argRegexStr)
					}

				} else {
					return "", errors.New("error reading arguments, unexpected '-' after argument " + args[i])
				}
			} else {
				return "", errors.New("error reading arguments, expected value after argument " + args[i])
			}
		}
	}
	return "", errors.New("error reading arguments, no arguments read")
}
