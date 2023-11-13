package utils

import "fmt"

func FromUtils() {
	fmt.Println("hello from utils")
}

//go:noescape
func Add(x, y int) int
func CallAsm() {
	a := Add(10, 20)
	fmt.Println("a ici :", a)
}

//go:noescape
func bpSyscall(callid uint16, argh ...uintptr) (errcode uint32)
func Syscall(callid uint16, argh ...uintptr) (errcode uint32, err error) {
	errcode = bpSyscall(callid, argh...)

	if errcode != 0 {
		err = fmt.Errorf("non-zero return from syscall")
	}
	return errcode, err
}
