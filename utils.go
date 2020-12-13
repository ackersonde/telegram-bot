package main

import (
	"log"
	"os/exec"
)

func getDeployFingerprint(deployCertFilePath string) string {
	out, err := exec.Command("/usr/bin/ssh-keygen", "-Lf", deployCertFilePath).Output()
	if err != nil {
		log.Println(err)
	}

	return string(out)
}
