package ui

import (
	"go-musetool/internal/language"
	"syscall"
	"unsafe"
)

const (
	mutexName = "Global\\GoMuseToolSingleInstanceMutex"
)

var (
	kernel32Mutex             = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexW          = kernel32Mutex.NewProc("CreateMutexW")
	procOpenMutexW            = kernel32Mutex.NewProc("OpenMutexW")
	procReleaseMutex          = kernel32Mutex.NewProc("ReleaseMutex")
	procCloseHandle           = kernel32Mutex.NewProc("CloseHandle")
	procGetLastError          = kernel32Mutex.NewProc("GetLastError")
	singleInstanceMutexHandle syscall.Handle
)

const (
	MUTEX_ALL_ACCESS     = 0x1F0001
	ERROR_ALREADY_EXISTS = 183
)

// CheckSingleInstance checks if another instance is already running
// Returns true if this is the first instance, false if already running
func CheckSingleInstance() bool {
	mutexNamePtr, err := syscall.UTF16PtrFromString(mutexName)
	if err != nil {
		return true // If error, allow to continue
	}

	// First try to open existing mutex
	existingHandle, _, _ := procOpenMutexW.Call(
		MUTEX_ALL_ACCESS,
		0,
		uintptr(unsafe.Pointer(mutexNamePtr)),
	)

	if existingHandle != 0 {
		// Mutex already exists, another instance is running
		procCloseHandle.Call(existingHandle)
		return false
	}

	// Mutex doesn't exist, create it
	handle, _, _ := procCreateMutexW.Call(
		0,
		0,
		uintptr(unsafe.Pointer(mutexNamePtr)),
	)

	if handle == 0 {
		return true // Failed to create mutex, allow to continue
	}

	singleInstanceMutexHandle = syscall.Handle(handle)
	return true
}

// ReleaseSingleInstance releases the mutex when application exits
func ReleaseSingleInstance() {
	if singleInstanceMutexHandle != 0 {
		procReleaseMutex.Call(uintptr(singleInstanceMutexHandle))
		procCloseHandle.Call(uintptr(singleInstanceMutexHandle))
		singleInstanceMutexHandle = 0
	}
}

// ShowAlreadyRunningDialog shows a dialog when instance is already running
func ShowAlreadyRunningDialog() {
	title := language.T().WindowTitle
	message := "程序已在运行中"

	// Show message box
	messagePtr, _ := syscall.UTF16PtrFromString(message)
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	const (
		MB_OK              = 0x00000000
		MB_ICONINFORMATION = 0x00000040
		MB_SYSTEMMODAL     = 0x00001000
	)

	user32 := syscall.NewLazyDLL("user32.dll")
	procMessageBoxW := user32.NewProc("MessageBoxW")

	procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		MB_OK|MB_ICONINFORMATION|MB_SYSTEMMODAL,
	)
}
