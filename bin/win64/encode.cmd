@echo off

SET "thisDir=%~dp0"
SET "encodeFile=%1"

if [%1]==[] (
	echo Please provide as parameter the file you wish to encode.
	echo The easiest way to do this is to drag and drop the input file onto this batch file.
	goto :end
)

echo Converting file to frames
"%thisDir%file2qr\cli.exe" -encode -encode-file=%encodeFile% -encode-folder-destination="%thisDir%qr-frames"
echo.

echo Converting frames to video
"%thisDir%ffmpeg\bin\ffmpeg.exe" -i "%thisDir%qr-frames\%%015d.png" -c:v libx264 -vf fps=25 -pix_fmt yuv420p "%thisDir%\result.mp4"
echo.

echo Cleaning up
del /F /Q "%thisDir%qr-frames"
rmdir "%thisDir%qr-frames"

:end
pause
