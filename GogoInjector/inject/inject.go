package inject

import (
	"fmt"
	"io/ioutil"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

func openProcess(processID uint32) (windows.Handle, error) {
	handle, err := windows.OpenProcess(
		windows.PROCESS_CREATE_THREAD|
			windows.PROCESS_QUERY_INFORMATION|
			windows.PROCESS_VM_OPERATION|
			windows.PROCESS_VM_WRITE|
			windows.PROCESS_VM_READ,
		false,
		processID)
	if err != nil {
		return 0, err
	}
	return handle, nil
}

func Mainfunc(PID uint32, dll_path string) error {
	fmt.Println("Hello from inject package", PID)

	// 1. Open dll to inject
	fmt.Println("\n------ 1. Open dll to inject")
	file, err := os.Open(dll_path)
	if err != nil {
		fmt.Println("Could not open dll to inject")
		return err
	}
	defer file.Close()

	// 2. Get dll size
	fmt.Println("\n------ 2. Get dll size")
	fi, err := file.Stat()
	if err != nil {
		fmt.Println("Could not get dll size")
		return err
	}
	dll_size := fi.Size()
	dwLength := *(*uint32)(unsafe.Pointer(&dll_size))
	fmt.Printf("The file is %d bytes long\n", dwLength)

	// 3. Get dll content
	fmt.Println("\n------ 3. Get dll content")
	dll_content, err := ioutil.ReadFile(dll_path)
	if err != nil {
		fmt.Println("Could not get dll content")
		return err
	}
	lpBuffer := uintptr(unsafe.Pointer(&dll_content[0]))
	fmt.Printf("File content type : %T\n", dll_content)
	fmt.Printf("lpBuffer type : %T\n", lpBuffer)
	fmt.Printf("File bytes of dll_content : %v\n", *(*byte)(unsafe.Pointer(lpBuffer + 1)))

	// 4. Get process handle
	fmt.Println("\n------ 4. Get dll handle")
	hProcess, err := openProcess(PID)
	//hProcess = uintptr(hProcess)
	if err != nil {
		fmt.Println("Could not get remote process handle")
		return err
	}
	fmt.Printf("Handle type : %T\n", hProcess)

	// 5. Load library remote
	fmt.Println("\n------ 5. Load library remote")
	LoadRemoteLibraryR(hProcess, lpBuffer, dwLength)
	return nil
}
