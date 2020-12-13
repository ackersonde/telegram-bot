package utils

import (
	"log"
	"os/exec"
)

// GetDeployFingerprint now commented
func GetDeployFingerprint(deployCertFilePath string) string {
	out, err := exec.Command("/usr/bin/ssh-keygen", "-Lf", deployCertFilePath).Output()
	if err != nil {
		log.Println(err)
	}

	return string(out)
}
