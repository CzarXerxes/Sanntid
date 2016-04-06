package main

/*
var state int //Is router in router mode or backup mode
const (
	router = iota
	backup
)
*/

func routerModuleInit() {
	spawnBackup()
}

func spawnBackup() {

}

func main() {
	routerModuleInit()
}
