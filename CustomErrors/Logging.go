package CustomErrors

import (
	"fmt"
	"os"
	"time"
)

func LogError(errorCode int, errorType string, exit bool, err error) {
	dt := time.Now()

	if errorType == "FATAL" {
		println(dt.Format("01-02-2006 15:04:05"), " FATAL ERROR: ", errorCode, err.Error())
		//log.Fatal("FATAL ERROR: ", errormsg, err.Error())
	} else if errorType == "WARNING" {
		println(dt.Format("01-02-2006 15:04:05"), " WARNING: ", errorCode, err.Error())
		//log.Print("WARNING: ", errormsg, err.Error())
	} else if errorType == "INFO" {
		println(dt.Format("01-02-2006 15:04:05"), " INFO: ", errorCode, err.Error())
		//log.Print("INFO: ", errormsg)
	} else {
		fmt.Println("Invalid error type sent to logError")
	}
	if exit {
		os.Exit(1)
	}
}
