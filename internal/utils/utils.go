package utils

import "github.com/prometheus/common/log"

//StopIfError is a convenient function to stop processing if there is any error
func StopIfError(err error) {
	if err != nil {
		log.Fatalf("error %v", err)
	}
}
