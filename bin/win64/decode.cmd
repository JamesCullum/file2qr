@echo off

SET thisDir=%~dp0
SET encodeFile=%1

if [%1]==[] (
	echo Please provide as parameter the movie file you wish to decode.
	echo The easiest way to do this is to drag and drop the input file onto this batch file.
	goto :end
)

echo Converting video to frames
mkdir "%thisDir%qr-frames"
"%thisDir%ffmpeg\bin\ffmpeg.exe" -i "%encodeFile%" "%thisDir%qr-frames\%%015d.png"
echo.

echo Converting frames to file
"%thisDir%file2qr\cli.exe" -decode -decode-folder="%thisDir%qr-frames" -decode-file-destination="%thisDir%\result.bin"
echo.

echo Cleaning up
del /F /Q "%thisDir%qr-frames"
rmdir "%thisDir%qr-frames"

:end
pause