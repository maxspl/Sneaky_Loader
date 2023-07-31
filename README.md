# GoRust_Sneaky_Syringe
Custom loader - Rust/Go Dll Reflective Injection - Only supports 32bits for the moment

# How to use it ?

1. Clone the repo
```
git clone <repo>
```

2. Compile the 32 bits dll
```
cd RustyLoader
rustup target add i686-pc-windows-msvc
cargo build --target=i686-pc-windows-msvc --release
```

3. Copy the compiled dll
```
Copy-Item .\target\i686-pc-windows-msvc\release\rusty_inject.dll -Destination ..\GogoInjector\
```

4. Compile the injector
```
cd GogoInjector
go get github.com/Binject/debug/pe
go get golang.org/x/sys/windows
$env:GOARCH="386"
go build .\inject.go
```

5. Test it

Spawn a 32 bit process and git its pid :
```
$process = Start-Process -FilePath "C:\Windows\SysWOW64\cmd.exe" -PassThru
$process.Id
```

Inject the dll :
```
.\inject.exe <pid>
```
![Alt text](/assets/injected.png)

# How to use the debug version ?

The directory debug_RustyLoader contains a verbose version of the RustyLoader. 
It works in a slightly different way :

	- It is an exe, not a dll. 
	- It operates as an injector and also a loader.
	- It reads the dll named debug_dll.dll in /dll_to_inject .
	- It loads the raw dll in the self process memory and perfoms all the loading operations as RustyLoader would do but with verbose printed logs. Once loaded, it calls dllmain.

Put the dll to be tested in the directory debug_RustyLoader\dll_to_inject and name it debug_dll.dll.
```
cd debug_RustyLoader
cargo run --target=i686-pc-windows-msvc
```
![Alt text](/assets/injected_verbose.png)