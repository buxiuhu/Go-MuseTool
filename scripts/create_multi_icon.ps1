# PowerShell script to create a multi-resolution icon from the existing 32x32 icon
# This script will create 16x16, 32x32, 48x48, and 256x256 versions

Add-Type -AssemblyName System.Drawing

$sourceIcon = "internal\assets\real_icon.ico"
$outputIcon = "internal\assets\real_icon_multi.ico"

Write-Host "Creating multi-resolution icon..."

try {
    # Load the source icon
    $icon = [System.Drawing.Icon]::new($sourceIcon)
    $bitmap = $icon.ToBitmap()
    
    Write-Host "Source icon: $($bitmap.Width)x$($bitmap.Height)"
    
    # Create bitmaps at different sizes
    $sizes = @(16, 32, 48, 256)
    $bitmaps = @()
    
    foreach ($size in $sizes) {
        $newBitmap = [System.Drawing.Bitmap]::new($size, $size)
        $graphics = [System.Drawing.Graphics]::FromImage($newBitmap)
        $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
        $graphics.DrawImage($bitmap, 0, 0, $size, $size)
        $graphics.Dispose()
        $bitmaps += $newBitmap
        Write-Host "  Created ${size}x${size} version"
    }
    
    # Save as multi-resolution icon
    $iconStream = [System.IO.FileStream]::new($outputIcon, [System.IO.FileMode]::Create)
    
    # Write ICO header
    $writer = [System.IO.BinaryWriter]::new($iconStream)
    $writer.Write([UInt16]0)  # Reserved
    $writer.Write([UInt16]1)  # Type (1 = ICO)
    $writer.Write([UInt16]$bitmaps.Count)  # Number of images
    
    # Calculate offsets
    $offset = 6 + ($bitmaps.Count * 16)  # Header + directory entries
    
    # Write directory entries and collect image data
    $imageDataList = @()
    foreach ($bmp in $bitmaps) {
        # Convert bitmap to PNG for better quality
        $ms = [System.IO.MemoryStream]::new()
        $bmp.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
        $imageData = $ms.ToArray()
        $ms.Dispose()
        
        # Write directory entry
        $width = if ($bmp.Width -eq 256) { 0 } else { $bmp.Width }
        $height = if ($bmp.Height -eq 256) { 0 } else { $bmp.Height }
        
        $writer.Write([byte]$width)
        $writer.Write([byte]$height)
        $writer.Write([byte]0)  # Color palette
        $writer.Write([byte]0)  # Reserved
        $writer.Write([UInt16]1)  # Color planes
        $writer.Write([UInt16]32)  # Bits per pixel
        $writer.Write([UInt32]$imageData.Length)  # Image size
        $writer.Write([UInt32]$offset)  # Image offset
        
        $imageDataList += $imageData
        $offset += $imageData.Length
    }
    
    # Write image data
    foreach ($imageData in $imageDataList) {
        $writer.Write($imageData)
    }
    
    $writer.Close()
    $iconStream.Close()
    
    # Clean up
    foreach ($bmp in $bitmaps) {
        $bmp.Dispose()
    }
    $bitmap.Dispose()
    $icon.Dispose()
    
    Write-Host "✓ Multi-resolution icon created: $outputIcon"
    
    # Verify the new icon
    $newIconBytes = [System.IO.File]::ReadAllBytes($outputIcon)
    $numImages = [BitConverter]::ToUInt16($newIconBytes, 4)
    Write-Host "  Contains $numImages images"
    
} catch {
    Write-Host "✗ Error: $_"
    Write-Host $_.Exception.StackTrace
    exit 1
}
