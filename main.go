package main
import ("fmt"
		 "os"
		 "strings" 
		 "bufio"
		 )



func openFile(path string) (string, error){
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
			content.WriteString(scanner.Text()+ "\n")
	}
	if err:= scanner.Err(); err != nil {
		return "", err
	}
	return content.String(), nil

}



func saveFile(path string, buf [][]rune) error {
    var b strings.Builder

    for i, line := range buf {
        b.WriteString(string(line))
        if i < len(buf)-1 {
            b.WriteByte('\n')
        }
    }

    return os.WriteFile(path, []byte(b.String()), 0644)
}


///Main Run Here
func main() {
	var filePath = "sample.json"
	
	
    // if len(os.Args) < 2 {
    //     fmt.Println("Usage: editor <filename>")
    //     return
    // }

    // filepath := os.Args[1]
    restore := enableRaw()
    defer restore()
		
	fmt.Print("\x1b[?1049h")
	defer fmt.Print("\x1b[?1049l")

    
    runQuickEditor(filePath)

}
