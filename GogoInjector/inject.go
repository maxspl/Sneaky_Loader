package main

import (
	"GogoInjector/inject"
	"fmt"
	"os"
	"strconv"
	"unsafe"
)

var dll_path = "rusty_inject.dll" //"acledit.dll"

func main() {

	if len(os.Args) < 2 {
		fmt.Println("No argument provided")
		return
	} else {
		PID, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("PID must be an int.")
			return
		}
		fmt.Println("Process ID to inject :", PID)

		dwPID := *(*uint32)(unsafe.Pointer(&PID))
		fmt.Printf("dwPID :%T\n", dwPID)

		err = inject.Mainfunc(dwPID, dll_path)
		if err != nil {
			fmt.Println("Error :", err)
			return
		}
	}

}
