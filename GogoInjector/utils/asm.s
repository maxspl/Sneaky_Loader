// +build windows,amd64

#include "textflag.h"

// func Add(x, y int) int
TEXT ·Add(SB), NOSPLIT, $0-24
    MOVQ x+0(FP), AX // Move x into AX.
    MOVQ y+8(FP), BX // Move y into BX.
    ADDQ BX, AX      // Add AX and BX, result in AX.
    MOVQ AX, ret+16(FP) // Move result to return value.
    RET


#define maxargs 16
//func Syscall(callid uint16, argh ...uintptr) (uint32, error)
TEXT ·bpSyscall(SB), $0-56
	XORQ AX,AX
	MOVW callid+0(FP), AX
	PUSHQ CX
	//put variadic size into CX
	MOVQ argh_len+16(FP),CX
	//put variadic pointer into SI
	MOVQ argh_base+8(FP),SI
	// SetLastError(0).
	MOVQ	0x30(GS), DI
	MOVL	$0, 0x68(DI)
	SUBQ	$(maxargs*8), SP	// room for args
	// Fast version, do not store args on the stack.
	CMPL	CX, $4
	JLE	loadregs
	// Check we have enough room for args.
	CMPL	CX, $maxargs
	JLE	2(PC)
	INT	$3			// not enough room -> crash
	// Copy args to the stack.
	MOVQ	SP, DI
	CLD
	REP; MOVSQ
	MOVQ	SP, SI
loadregs:
	//move the stack pointer????? why????
	SUBQ	$8, SP
	// Load first 4 args into correspondent registers.
	MOVQ	0(SI), CX
	MOVQ	8(SI), DX
	MOVQ	16(SI), R8
	MOVQ	24(SI), R9
	// Floating point arguments are passed in the XMM
	// registers. Set them here in case any of the arguments
	// are floating point values. For details see
	//	https://msdn.microsoft.com/en-us/library/zthk2dkh.aspx
	MOVQ	CX, X0
	MOVQ	DX, X1
	MOVQ	R8, X2
	MOVQ	R9, X3
	//MOVW callid+0(FP), AX
	MOVQ CX, R10
	SYSCALL
	ADDQ	$((maxargs+1)*8), SP
	// Return result.
	POPQ	CX
	MOVL	AX, errcode+32(FP)
	RET


//func GetPEB() uintptr
TEXT ·GetPEB(SB), $0-8
     MOVQ 	0x60(GS), AX
     MOVQ	AX, ret+0(FP)
     RET


// func execIndirectSyscall(ssn uint16, trampoline uintptr, argh ...uintptr) uint32
TEXT ·execIndirectSyscall(SB),NOSPLIT, $0-40
    XORQ    AX, AX
    MOVW    ssn+0(FP), AX
	
    XORQ    R11, R11
    MOVQ    trampoline+8(FP), R11
	
    PUSHQ   CX
	
    //put variadic pointer into SI
    MOVQ    argh_base+16(FP),SI

    //put variadic size into CX
    MOVQ    argh_len+24(FP),CX
	
    // SetLastError(0).
    MOVQ    0x30(GS), DI
    MOVL    $0, 0x68(DI)

    // room for args
    SUBQ    $(maxargs*8), SP	

    //no parameters, special case
    CMPL    CX, $0
    JLE     jumpcall
	
    // Fast version, do not store args on the stack.
    CMPL    CX, $4
    JLE	    loadregs

    // Check we have enough room for args.
    CMPL    CX, $maxargs
    JLE	    2(PC)

    // not enough room -> crash
    INT	    $3			

    // Copy args to the stack.
    MOVQ    SP, DI
    CLD
    REP; MOVSQ
    MOVQ    SP, SI
	
loadregs:

    // Load first 4 args into correspondent registers.
    MOVQ	0(SI), CX
    MOVQ	8(SI), DX
    MOVQ	16(SI), R8
    MOVQ	24(SI), R9
	
    // Floating point arguments are passed in the XMM registers
    // Set them here in case any of the arguments are floating point values. 
    // For details see: https://msdn.microsoft.com/en-us/library/zthk2dkh.aspx
    MOVQ	CX, X0
    MOVQ	DX, X1
    MOVQ	R8, X2
    MOVQ	R9, X3
	
jumpcall:
    MOVQ    CX, R10

    //jump to syscall;ret gadget address instead of direct syscall
    CALL    R11

    ADDQ	$((maxargs)*8), SP

    // Return result
    POPQ	CX
    MOVL	AX, errcode+40(FP)
    RET
