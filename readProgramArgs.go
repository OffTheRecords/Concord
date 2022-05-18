package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type runTimeArgs struct {
	valid               bool
	logLevel            string
	dbUserMongo         string
	dbPassMongo         string
	dbPortMongo         string
	dbHostMongo         string
	dbNameMongo         string
	redisGlobalHostAddr string
	redisGlobalHostPort string
	redisGlobalPassword string
}

func readRunTimeArgs() runTimeArgs {

	//Get runtime args
	var programArgs runTimeArgs

	//Get log level
	logLevel, err := readStringArg("logging", "\\Adebug|error|warning|info$")
	if err != nil {
		fmt.Print(err.Error() + "\n")
		programArgs.valid = false
		return programArgs
	}
	programArgs.logLevel = logLevel

	//Get Mongo Database user
	dbUserMongo, err := readStringArg("dbUserMongo", "\\A.{4,128}$")
	if err != nil {
		fmt.Print(err.Error() + "\n")
		programArgs.valid = false
		return programArgs
	}
	programArgs.dbUserMongo = dbUserMongo

	//Get Mongo Database password
	dbPassMongo, err := readStringArg("dbPassMongo", "\\A.{8,128}$")
	if err != nil {
		fmt.Print(err.Error() + "\n")
		programArgs.valid = false
		return programArgs
	}
	programArgs.dbPassMongo = dbPassMongo

	//Get Mongo server port
	dbPortMongo, err := readStringArg("dbPortMongo", "\\A[0-9]{1,5}$")
	if err != nil {
		fmt.Print(err.Error() + "\n")
		programArgs.valid = false
		return programArgs
	}
	programArgs.dbPortMongo = dbPortMongo

	//Get Mongo database host address
	dbHostMongo, err := readStringArg("dbHostMongo", "\\A.{1,128}$")
	if err != nil {
		fmt.Print(err.Error() + "\n")
		programArgs.valid = false
		return programArgs
	}
	programArgs.dbHostMongo = dbHostMongo

	//Get Mongo database name
	dbNameMongo, err := readStringArg("dbNameMongo", "\\A.{1,64}$")
	if err != nil {
		fmt.Print(err.Error() + "\n")
		programArgs.valid = false
		return programArgs
	}
	programArgs.dbNameMongo = dbNameMongo

	//Get global redis host address
	redisGlobalHostAddr, err := readStringArg("redisGlobalHostAddr", "\\A.{1,128}$")
	if err != nil {
		fmt.Print(err.Error() + "\n")
		programArgs.valid = false
		return programArgs
	}
	programArgs.redisGlobalHostAddr = redisGlobalHostAddr

	//Get global redis port
	redisGlobalHostPort, err := readStringArg("redisGlobalHostPort", "\\A[0-9]{1,5}$")
	if err != nil {
		fmt.Print(err.Error() + "\n")
		programArgs.valid = false
		return programArgs
	}
	programArgs.redisGlobalHostPort = redisGlobalHostPort

	//Get global redis password
	redisGlobalHostPassword, err := readStringArg("redisGlobalHostPassword", "\\A.{1,128}$")
	if err != nil {
		fmt.Print(err.Error() + "\n")
		programArgs.valid = false
		return programArgs
	}
	programArgs.redisGlobalHostPort = redisGlobalHostPassword

	programArgs.valid = true

	return programArgs
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

	if len(args) == 0 {
		return "", errors.New("error reading arguments, no arguments read")
	} else {
		return "", errors.New("error reading arguments, argument " + argName + " not found")
	}
}
