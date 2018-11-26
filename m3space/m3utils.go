package m3space

import (
	"encoding/csv"
	"log"
	"os"
)

func changeToDocsGeneratedDir() {
	changeToDocsSubdir("generated")
}

func changeToDocsDataDir() {
	changeToDocsSubdir("data")
}

func changeToDocsSubdir(subDir string) {
	if _, err := os.Stat("docs"); !os.IsNotExist(err) {
		exitOnError(os.Chdir("docs"))
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			exitOnError(os.Mkdir(subDir, os.ModePerm))
		}
		exitOnError(os.Chdir(subDir))
	}
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

/*func writeNextBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	exitOnError(err)
}
*/

func closeFile(file *os.File) {
	exitOnError(file.Close())
}

func writeNextString(file *os.File, text string) {
	_, err := file.WriteString(text)
	exitOnError(err)
}

func writeAll(writer *csv.Writer, records [][]string) {
	exitOnError(writer.WriteAll(records))
}

func write(writer *csv.Writer, records []string) {
	exitOnError(writer.Write(records))
}
