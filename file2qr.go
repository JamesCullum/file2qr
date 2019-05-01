package file2qr

import (
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"io/ioutil"

	"image"
	"image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// Converter is the instance of this converter.
// You can directly manipulate the available parameters to modify settings.
type Converter struct {
	NumThread         int
	QRChunkSize       int64
	QRSideLength      int
	FilenamePadLength int
	FilePrefix        string
}

// New creates a converter instance with default settings.
func New() *Converter {
	return &Converter{
		NumThread:         25,
		QRChunkSize:       2950,
		QRSideLength:      1480,
		FilenamePadLength: 15,
		FilePrefix:        "",
	}
}

// Encode reads the input file and converts it to QR code images in the destination folder.
// It writes the current progress percentage into the provided progress reference integer.
func (c *Converter) Encode(inputFilepath, outputFolderpath string, progress *int) error {
	// Open up input file. Will be read in batches to reduce memory impact.
	inputFile, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	fileStat, err := inputFile.Stat()
	if err != nil {
		return err
	}

	// Create folder if it does not exist
	os.Mkdir(outputFolderpath, 0644)

	// Init encoder
	encoder := qrcode.NewQRCodeWriter()
	encodeHints := map[gozxing.EncodeHintType]interface{}{gozxing.EncodeHintType_ERROR_CORRECTION: "L", gozxing.EncodeHintType_CHARACTER_SET: "ASCII"}

	*progress = 0
	maxI := int(math.Floor(float64(fileStat.Size())/float64(c.QRChunkSize))) + 1
	readBuffer := make([]byte, c.QRChunkSize)

	sem := make(chan bool, c.NumThread)
	for i := 1; i <= maxI; i++ {
		sem <- true

		// Read chunk
		readBytes, err := inputFile.Read(readBuffer)
		if err != nil {
			return err
		}

		// Hand over to thread to encode and write to file
		go func(i int, chunk string) {
			qrImage, err := encoder.Encode(chunk, gozxing.BarcodeFormat_QR_CODE, c.QRSideLength, c.QRSideLength, encodeHints)
			if err != nil {
				panic(err)
			}

			filename := c.FilePrefix + lpad(strconv.Itoa(i), "0", c.FilenamePadLength) + ".png"
			f, err := os.Create(outputFolderpath + string(os.PathSeparator) + filename)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			if err := png.Encode(f, qrImage); err != nil {
				panic(err)
			}

			// Update progress
			tempProgress := int((i * 100) / maxI)
			if tempProgress > *progress {
				*progress = tempProgress
			}

			<-sem
		}(i, string(readBuffer[:readBytes]))
	}

	// Wait for all threads
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

// Decode reads qr codes from a source folder, decodes them and writes the result in the destination file.
// The source folder should only contain images that belong to one file.
// It writes the current progress percentage into the provided progress reference integer.
func (c *Converter) Decode(inputFolderpath, outputFilepath string, progress *int) error {
	// Init decoder twice with different settings as fallback.
	// See https://github.com/makiuchi-d/gozxing/issues/14#issuecomment-486279433 for details.
	decoder := qrcode.NewQRCodeReader()
	decodeHints := make(map[gozxing.DecodeHintType]interface{})
	decodeHints[gozxing.DecodeHintType_PURE_BARCODE] = true
	decodeHints2 := make(map[gozxing.DecodeHintType]interface{})

	outputFile, err := os.Create(outputFilepath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Read files from folder to check how many are available
	files, err := ioutil.ReadDir(inputFolderpath)
	if err != nil {
		return err
	}
	lock := sync.Mutex{}

	iStart := 1
	iEnd := len(files)
	resultAssemble := make(map[int][]byte, iEnd)
	sem := make(chan bool, c.NumThread+1)

	// To reduce memory, we start a thread to stream memory to disk.
	// It also needs to take a thread, so that the last files are written before the program exits.
	go func(iStart, iEnd int, progress *int, outputFile *os.File, resultAssemble map[int][]byte, lock *sync.Mutex, sem chan bool) {
		sem <- true

		streamToFile(iStart, iEnd, progress, outputFile, resultAssemble, lock, sem)
	}(iStart, iEnd, progress, outputFile, resultAssemble, &lock, sem)

	// Iterating over files
	for i := iStart; i <= iEnd; i++ {
		sem <- true

		// Reading and decoding in threads
		go func(i int, resultAssemble map[int][]byte, lock *sync.Mutex) {
			filename := c.FilePrefix + lpad(strconv.Itoa(i), "0", c.FilenamePadLength) + ".png"
			f, err := os.Open(inputFolderpath + string(os.PathSeparator) + filename)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				panic(err)
			}

			bmp, err := gozxing.NewBinaryBitmapFromImage(img)
			if err != nil {
				panic(err)
			}

			decodeResult, err := decoder.Decode(bmp, decodeHints)
			if err != nil {
				decodeResult, err = decoder.Decode(bmp, decodeHints2)

				if err != nil {
					panic(err)
				}
			}

			lock.Lock()
			resultAssemble[i] = []byte(decodeResult.GetText())
			lock.Unlock()

			<-sem
		}(i, resultAssemble, &lock)
	}

	// Wait for all threads
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

// streamToFile writes content continuously from memory to disk to free memory
// while maintaining the order when writing to file
func streamToFile(currentI, endI int, progress *int, outputFile *os.File, resultAssemble map[int][]byte, lock *sync.Mutex, sem chan bool) {
	for i := currentI; i <= endI; i++ {
		if chunk, ok := resultAssemble[i]; ok {
			tempProgress := int((currentI * 100) / endI)
			if tempProgress > *progress {
				*progress = tempProgress
			}

			// Write to disk
			outputFile.Write(chunk)

			// Free memory
			lock.Lock()
			resultAssemble[i] = nil
			delete(resultAssemble, i)
			lock.Unlock()
		} else {
			// Not ready yet, sleep & restart
			time.Sleep(2 * time.Second)
			streamToFile(i, endI, progress, outputFile, resultAssemble, lock, sem)
			return
		}
	}
	*progress = 100
	<-sem
}

// lpad pads a string to a certain length using a provided pad character
// Original credits: https://stackoverflow.com/a/45456649/1424378
func lpad(s string, pad string, plength int) string {
	for i := len(s); i < plength; i++ {
		s = pad + s
	}
	return s
}
