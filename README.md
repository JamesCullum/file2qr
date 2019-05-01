# file2qr

[![travis](https://travis-ci.org/JamesCullum/file2qr.svg?branch=master)](https://travis-ci.org/JamesCullum/file2qr) 
[![go report card](https://goreportcard.com/badge/github.com/JamesCullum/file2qr)](https://goreportcard.com/report/github.com/JamesCullum/file2qr) 
[![coverage](https://img.shields.io/badge/coverage-86%25-brightgreen.svg)](https://gocover.io/github.com/JamesCullum/file2qr)
[![godocs](https://godoc.org/github.com/JamesCullum/file2qr?status.svg)](https://godoc.org/github.com/JamesCullum/file2qr) 

## Install

To use it in your script, pull it as below and include it in your script.

```
go get github.com/JamesCullum/file2qr
```

If you want to use it as standalone client including the CLI, download the newest release for your operating system.

## Usage

### Basic usage

```golang
// Import at top of script
import "github.com/JamesCullum/file2qr"

// Initiate handler with default settings
handler := file2qr.New()

// Methods update this value as percentage, so that you can display it in a progressbar
current_progress := 0

// Convert `input.bin` to qr codes in the folder `frames`
// Download the newest release to get a version packaged with ffmpeg to create a video out of the frames
err := handler.Encode("./input.bin", "./frames", &current_progress)

// Convert qr codes in the folder `frames` to `output.bin`
// Download the newest release to get a version packaged with ffmpeg to extract each code frame out of a video
err := handler.Decode("./frames", "./output.bin", &current_progress)
```

### Use as standalone converter

Download the newest release for your operationg system.
It will contain a compiled CLI and packages ffmpeg for video en/decoding.

It contains shell scripts for encoding and decoding a file to qr codes and a video file of it.
Execute it in a console to get further instructions.

On Windows, you can simply drag and drop the file you wish to encode or the video file you wish to decode onto the batch file.
This will create a `result.bin` (decoding) or `result.mp4` folder (encoding) in the CLI directory.

## Contributing

Pull requests are welcome. Feel free to...

- Revise documentation
- Add new features
- Fix bugs
- Suggest improvements

## Thanks

Thanks [@scholz](https://github.com/schollz) for the inspiration to create a simple and good Go Github project

## License

MIT
