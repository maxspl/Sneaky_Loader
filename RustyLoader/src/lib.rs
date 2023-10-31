#![no_std] // * Can be commented
#![no_main]

use core::{ffi::c_void, ptr::null_mut, slice::from_raw_parts, mem::{transmute, size_of}, arch::asm};
use ntapi::winapi::shared::minwindef::{DWORD, LPVOID, HINSTANCE, BOOL, TRUE};
use windows_sys::Win32::{System::{SystemServices::DLL_PROCESS_ATTACH}, UI::WindowsAndMessaging::MessageBoxA};
mod loader;

#[cfg(not(test))] // * Can be commented
#[panic_handler] // * Can be commented
fn panic(_info: &core::panic::PanicInfo) -> ! { loop {} } // * Can be commented

#[export_name = "_fltused"]
static _FLTUSED: i32 = 0;

#[no_mangle]
#[allow(non_snake_case)]
pub unsafe extern "system" fn _DllMainCRTStartup( // * can be replaced with DllMain
    _module: HINSTANCE,
    call_reason: DWORD,
    _reserved: LPVOID,
) -> BOOL {
    if call_reason == DLL_PROCESS_ATTACH {
        // Cleanup RWX region
        // VirtualFree(_reserved as _, 0, MEM_RELEASE);
        #[cfg(target_pointer_width = "32")]
        let message = "Touran toute neuve has been injected (x32) !\0";

        #[cfg(target_pointer_width = "64")]
        let message = "Touran toute neuve has been injected (x64) !\0";
        MessageBoxA(
            0 as _,
            message.as_ptr() as _,
            "Vroom vroom\0".as_ptr() as _,
            0x0,
        );


        TRUE
    } else {
        TRUE
    }
}