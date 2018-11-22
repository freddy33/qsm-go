package main

import (
	"os"
	"fmt"
	"github.com/freddy33/qsm-go/playgl"
	"github.com/freddy33/qsm-go/m3space"
)

func main() {
	c := "play1"
	if len(os.Args) > 1 {
		c = os.Args[1]
	}
	fmt.Println("Executing", c)
	switch c {
	case "play1":
		playgl.DisplayPlay1()
	case "writeTables":
		changeToDocsDir()
		m3space.WriteAllTrioTable()
		m3space.WriteTrioConnectionsTable()
		m3space.WriteAllConnectionDetails()
	default:
		fmt.Println("The param",c,"unknown")
	}
	fmt.Println("Finished Executing", c)
}

func changeToDocsDir() {
	if _, err := os.Stat("docs"); !os.IsNotExist(err) {
		os.Chdir("docs")
		if _, err := os.Stat("generated"); os.IsNotExist(err) {
			os.Mkdir("generated", os.ModePerm)
		}
		os.Chdir("generated")
	}
}
