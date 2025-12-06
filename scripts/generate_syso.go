package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Check if icon file exists
	iconPath := "internal/assets/real_icon.ico"
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		fmt.Printf("Error: Icon file not found: %s\n", iconPath)
		os.Exit(1)
	}

	// Remove old syso file
	os.Remove("GoMuseTool.syso")

	// Generate syso with rsrc
	cmd := exec.Command("rsrc",
		"-ico", iconPath,
		"-o", "GoMuseTool.syso",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running rsrc: %v\n", err)
		fmt.Printf("Output: %s\n", output)
		os.Exit(1)
	}

	fmt.Println("✓ Successfully generated GoMuseTool.syso")

	// Check file size
	info, _ := os.Stat("GoMuseTool.syso")
	fmt.Printf("  File size: %d bytes\n", info.Size())
}
