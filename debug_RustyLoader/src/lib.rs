use std::fs;
use core::{ffi::c_void, ptr::null_mut, slice::from_raw_parts, mem::{transmute, size_of}, arch::asm};
use ntapi::winapi::shared::minwindef::{DWORD, LPVOID, HINSTANCE, BOOL, TRUE};
use windows_sys::Win32::{System::{SystemServices::DLL_PROCESS_ATTACH}, UI::WindowsAndMessaging::MessageBoxA};
mod loader;


#[no_mangle]
#[allow(non_snake_case)]
pub unsafe extern "system" fn _DllMainCRTStartup(
    _module: HINSTANCE,
    call_reason: DWORD,
    _reserved: LPVOID,
) -> BOOL {
    if call_reason == DLL_PROCESS_ATTACH {
        // Cleanup RWX region
        // VirtualFree(_reserved as _, 0, MEM_RELEASE);
        MessageBoxA(
            0 as _,
            "Rust DLL injected!\0".as_ptr() as _,
            "Rust DLL\0".as_ptr() as _,
            0x0,
        );


        TRUE
    } else {
        TRUE
    }
}