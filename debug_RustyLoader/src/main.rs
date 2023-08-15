use std::fs;
use core::{ffi::c_void, ptr::null_mut, slice::from_raw_parts, mem::{transmute, size_of}, arch::asm};
use ntapi::winapi::shared::minwindef::{DWORD, LPVOID, HINSTANCE, BOOL, TRUE};
use windows_sys::Win32::{System::{SystemServices::DLL_PROCESS_ATTACH}, UI::WindowsAndMessaging::MessageBoxA};
mod loader;

fn main() {

    loader::hello_from_lib();
    #[cfg(target_pointer_width = "32")]
    let data = fs::read("dll_to_inject\\rusty_inject_32.dll").unwrap();

    #[cfg(target_pointer_width = "64")]
    let data = fs::read("dll_to_inject\\rusty_inject.dll").unwrap();

    let pointer = data.as_ptr();
    println!("Pointer to the beginning of data: {:?}", pointer);
    let void_ptr: *mut c_void = pointer as *mut c_void;
    unsafe {
        loader::ReflectiveLoader(void_ptr);
        
    }
}

