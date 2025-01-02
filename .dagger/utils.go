package main

import (
	"fmt"
)

func getVersion(version, tag, commit string, uncommited bool) string {
	if len(version) == 0 {
		state := "clean"
		if uncommited {
			state = "dirty"
		}

		if state == "clean" && len(tag) > 0 {
			version = tag
		} else {
			version = fmt.Sprintf("%s-%s", commit, state)
		}
	}

	return version
}
