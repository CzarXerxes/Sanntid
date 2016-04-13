//Hello
//Welcome to this elevator project
//This file opens all other files
package main

import (
	"os/exec"
)

/*
sshConfig := &ssh.ClientConfig{
	User: "your_user_name",
	Auth: []ssh.AuthMethod{
		ssh.Password("your_password")
	},
}
*/

const IP1 = "129.241.187.148" //Start router on this IP
const IP2 = "129.241.187.147"
const IP3 = "129.241.187.142"

func main() {
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "export GOPATH=~/Desktop/Heis")
	_ = cmd.Run()
	cmd = exec.Command("gnome-terminal", "-x", "sh", "-c", "go install driver")
	_ = cmd.Run()
	cmd = exec.Command("gnome-terminal", "-x", "sh", "-c", "go install user")
	_ = cmd.Run()
	cmd = exec.Command("gnome-terminal", "-x", "sh", "-c", "go install elevator")
	_ = cmd.Run()
	cmd = exec.Command("gnome-terminal", "-x", "sh", "-c", "go install control")
	_ = cmd.Run()
	cmd = exec.Command("gnome-terminal", "-x", "sh", "-c", "go install network")
	_ = cmd.Run()
	cmd = exec.Command("gnome-terminal", "-x", "sh", "-c", "go run RouterModule.go")
	_ = cmd.Run()
	cmd = exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go")
	_ = cmd.Run()
	//Do these remotely
	cmd = exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go")
	_ = cmd.Run()
	cmd = exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go")
	_ = cmd.Run()
}
