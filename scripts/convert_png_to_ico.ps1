# PNG to ICO Converter
# This script converts the PNG file to a proper multi-resolution ICO file

Add-Type -AssemblyName System.Drawing

$sourcePng = "icons\GOMuseTool.ico"  # Actually a PNG file
$outputIco = "icons\GoMuseTool_converted.ico"

Write-Host "Converting PNG to ICO format..." -ForegroundColor Cyan
Write-Host ""

try {
    # Load the PNG image
    $image = [System.Drawing.Image]::FromFile((Resolve-Path $sourcePng).Path)
    Write-Host "Source image loaded: $($image.Width)x$($image.Height)" -ForegroundColor Green
    
    # Create bitmaps at different sizes
    $sizes = @(16, 32, 48, 256)
    $icons = @()
    
    foreach ($size in $sizes) {
        Write-Host "Creating ${size}x${size} version..." -ForegroundColor Yellow
        
        $bitmap = New-Object System.Drawing.Bitmap($size, $size)
        $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
        $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
        $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
        $graphics.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
        $graphics.DrawImage($image, 0, 0, $size, $size)
        $graphics.Dispose()
        
        $icons += $bitmap
    }
    
    # Create ICO file manually
    $stream = [System.IO.FileStream]::new($outputIco, [System.IO.FileMode]::Create)
    $writer = [System.IO.BinaryWriter]::new($stream)
    
    # ICO Header
    $writer.Write([UInt16]0)  # Reserved
    $writer.Write([UInt16]1)  # Type (1 = ICO)
    $writer.Write([UInt16]$icons.Count)  # Number of images
    
    # Calculate offsets
    $offset = 6 + ($icons.Count * 16)
    $imageData = @()
    
    # Convert each bitmap to PNG and collect data
    foreach ($bitmap in $icons) {
        $ms = New-Object System.IO.MemoryStream
        $bitmap.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
        $data = $ms.ToArray()
        $ms.Dispose()
        $imageData += , $data
    }
    
    # Write directory entries
    for ($i = 0; $i -lt $icons.Count; $i++) {
        $bitmap = $icons[$i]
        $data = $imageData[$i]
        
        $width = if ($bitmap.Width -eq 256) { 0 } else { $bitmap.Width }
        $height = if ($bitmap.Height -eq 256) { 0 } else { $bitmap.Height }
        
        $writer.Write([byte]$width)
        $writer.Write([byte]$height)
        $writer.Write([byte]0)  # Color palette
        $writer.Write([byte]0)  # Reserved
        $writer.Write([UInt16]1)  # Color planes
        $writer.Write([UInt16]32)  # Bits per pixel
        $writer.Write([UInt32]$data.Length)  # Image size
        $writer.Write([UInt32]$offset)  # Image offset
        
        $offset += $data.Length
    }
    
    # Write image data
    foreach ($data in $imageData) {
        $writer.Write($data)
    }
    
    $writer.Close()
    $stream.Close()
    
    # Clean up
    foreach ($bitmap in $icons) {
        $bitmap.Dispose()
    }
    $image.Dispose()
    
    Write-Host ""
    Write-Host "✓ Conversion successful!" -ForegroundColor Green
    Write-Host "  Output: $outputIco" -ForegroundColor Gray
    
    # Verify the output
    $outBytes = [System.IO.File]::ReadAllBytes($outputIco)
    $numImages = [BitConverter]::ToUInt16($outBytes, 4)
    Write-Host "  Contains $numImages images" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "  1. Rename the converted file to replace the original" -ForegroundColor White
    Write-Host "  2. Or update your configuration to use the new file" -ForegroundColor White
    
}
catch {
    Write-Host ""
    Write-Host "✗ Conversion failed: $_" -ForegroundColor Red
    Write-Host $_.Exception.StackTrace -ForegroundColor Gray
    exit 1
}
