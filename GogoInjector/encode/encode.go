package encode

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func Encode(path string) error {
	var dll_content []byte

	//get dll content
	fmt.Println("[INFO] + Get dll content")
	dll_content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("[ERROR] - Could not get dll content")
		return err
	}
	counter := 0
	dll_size := len(dll_content)
	encoded_dll := dll_content
	key := byte(0xde)
	fmt.Println("[INFO] + Done", dll_content[0:4])
	// xor the content of the dll
	for counter != dll_size {
		encoded_dll[counter] = dll_content[counter] ^ key
		counter++
	}

	// write output to file
	output_dll := strings.Split(path, ".")[0] + ".enc"
	fmt.Println("[INFO] + Writing enc content to file : ", output_dll)
	err = os.WriteFile(output_dll, encoded_dll, 0644)
	if err != nil {
		fmt.Println("[ERROR] - Error writing to file:", err)
	}
	fmt.Println("[INFO] + Done")

	fmt.Println("[INFO] + Done", encoded_dll[0:4])

	return nil
}
