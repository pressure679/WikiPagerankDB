package main
import (
	"fmt"
	"io/ioutil"
	"os"
	"github.com/Obaied/rake"
)
type Stuff struct {
	txt string
	pagerank map[string]float64
	rake float64
	// avg map[string]float64
}
func main() {
	data := make(map[string]*Stuff)
	defer recover()
	files, err := listDir("files")
	if err != nil { panic(err) }
	for _, file := range files {
		data[file] = &Stuff{}
		data[file].txt, err = fileReader(file)
		if err != nil { panic(err) }
		cands := rake.RunRake(data[file].txt)
		for _, cand := range cands {
			fmt.Println(cand.Value, "-", cand.Key)
		}
	}
}
func listDir(dir string) (files []string, err error) {
	osFileInfo, err := ioutil.ReadDir(dir)
	if err != nil { return nil, err }
	for _, file := range osFileInfo {
		if !file.IsDir() {
			files = append(files, file.Name())
		}
	}
	return files, nil
}
func fileReader(file string) (txt string, err error) {
	osFile, err := os.Open("files/" + file)
	if err != nil { return "", err }
	defer osFile.Close()
	bytes, err := ioutil.ReadFile(osFile.Name())
	if err != nil { return "", err }
	return string(bytes), err
}
