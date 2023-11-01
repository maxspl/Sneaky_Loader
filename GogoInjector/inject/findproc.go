package inject

import (
	"fmt"

	ps "github.com/mitchellh/go-ps"
)

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func Findproc() int {
	processList, err := ps.Processes()
	if err != nil {
		fmt.Println("ps.Processes() Failed")
		return 0
	}
	for x := range processList {
		var process ps.Process
		process = processList[x]
		plist := []string{"OneDrive.exe", "explorer.exe", "svchost.exe"}
		if Contains(plist, process.Executable()) {
			fmt.Printf("%d\t%s\n", process.Pid(), process.Executable())
			pidToRetrun := process.Pid()
			return pidToRetrun
		}

	}
	return 0
	// do os.* stuff on the pid
	// onedrive.exe explorer.exe conhost.exe

}
