package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// https://gist.github.com/ammarshah/f5c2624d767f91a7cbdc4e54db8dd0bf
// https://gist.github.com/tbrianjones/5992856

func main() {
	files, _ := filepath.Glob("data/**/*.txt")
	domains := make(map[string]int)

	for index := range files {
		fileName := files[index]
		fmt.Println(fileName)
		file, _ := os.Open(fileName)
		fscanner := bufio.NewScanner(file)
		for fscanner.Scan() {
			// fmt.Println(fscanner.Text())
			txt := fscanner.Text()
			domains[txt] += 1
		}
	}

	outF, err := os.Create("data/domains.txt")
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer outF.Close()

	for key, _ := range domains {
		outF.WriteString(fmt.Sprintf("%s\n", key))
	}
	outF.Sync()

	fmt.Println(len(domains))
}
