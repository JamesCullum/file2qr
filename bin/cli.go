package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/JamesCullum/file2qr"
	"github.com/schollz/progressbar"
)

func main() {
	dir, _ := os.Getwd()

	// Console arguments
	flagEncode := flag.Bool("encode", false, "Encode a file")
	flagEncodeSrc := flag.String("encode-file", dir+string(os.PathSeparator)+`input.mp4`, "File to encode")
	flagEncodeDst := flag.String("encode-folder-destination", dir+string(os.PathSeparator)+`frames-encoded`, "Folder to write the qr codes into")

	flagDecode := flag.Bool("decode", false, "Decode a file")
	flagDecodeSrc := flag.String("decode-folder", dir+string(os.PathSeparator)+`frames-encoded`, "Folder with qr codes to decode")
	flagDecodeDst := flag.String("decode-file-destination", dir+string(os.PathSeparator)+`output.mp4`, "File where all decoded codes should be saved as")

	flag.Parse()
	if *flagEncode == false && *flagDecode == false {
		flag.PrintDefaults()
		return
	}

	// Init progressbar
	currentProgress := 0
	progress := progressbar.New(100)
	go updateProgressbar(&currentProgress, progress)

	// Init handler
	handler := file2qr.New()

	// Encode
	if *flagEncode == true {
		err := handler.Encode(*flagEncodeSrc, *flagEncodeDst, &currentProgress)
		if err != nil {
			log.Panic("Encoding failed: ", err)
		}
		currentProgress = 100
		progress.Finish()
	}

	// Reset progressbar
	if *flagEncode == true && *flagDecode == true {
		currentProgress = 0
		progress.Reset()
		go updateProgressbar(&currentProgress, progress)
	}

	// Decode
	if *flagDecode == true {
		err := handler.Decode(*flagDecodeSrc, *flagDecodeDst, &currentProgress)
		if err != nil {
			log.Panic("Decoding failed: ", err)
		}
		currentProgress = 100
		progress.Finish()
	}
}

func updateProgressbar(currentProgress *int, progress *progressbar.ProgressBar) {
	for range time.Tick(1 * time.Second) {
		progress.Set(*currentProgress)

		if *currentProgress >= 100 {
			progress.Finish()
			return
		}
	}
}
