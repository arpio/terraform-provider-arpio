package arpio

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

func WaitForDebuggerAttach() error {
	envVar := ArpioDebugWaitEnv
	durationStr, ok := os.LookupEnv(envVar)
	if ok {
		durationInt, err := strconv.Atoi(durationStr)
		if err != nil {
			return fmt.Errorf("Environment variable %s is set to an "+
				"invalid wait duration %q; specify an integer number of seconds\n",
				envVar, durationStr)
		}

		log.Printf("[INFO] Sleeping %d seconds for debugger to attach...\n", durationInt)
		time.Sleep(time.Duration(durationInt * 1000000000))
		log.Println("[INFO] Continuing after debugger sleep")
	}

	return nil
}
