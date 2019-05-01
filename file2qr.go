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

type converter struct {
	NumThread         int
	QRChunkSize       int64
	QRSideLength      int
	FilenamePadLength int
	FilePrefix        string
}

// New creates a converter instance with default settings.
// You can directly manipulate the available parameters to modify settings.
func New() *converter {
	return &converter{
		NumThread:         25,
		QRChunkSize:       2950,
		QRSideLength:      1480,
		FilenamePadLength: 15,
		FilePrefix:        "",
	}
}

// Encode reads the input file and converts it to QR code images in the destination folder.
// It writes the current progress percentage into the provided progress reference integer.
func (c *converter) Encode(input_filepath, output_folderpath string, progress *int) error {
	// Open up input file. Will be read in batches to reduce memory impact.
	input_file, err := os.Open(input_filepath)
	if err != nil {
		return err
	}
	defer input_file.Close()

	file_stat, err := input_file.Stat()
	if err != nil {
		return err
	}

	// Create folder if it does not exist
	os.Mkdir(output_folderpath, 0644)

	// Init encoder
	encoder := qrcode.NewQRCodeWriter()
	encode_hints := map[gozxing.EncodeHintType]interface{}{gozxing.EncodeHintType_ERROR_CORRECTION: "L", gozxing.EncodeHintType_CHARACTER_SET: "ASCII"}

	*progress = 0
	max_i := int(math.Floor(float64(file_stat.Size())/float64(c.QRChunkSize))) + 1
	read_buffer := make([]byte, c.QRChunkSize)

	sem := make(chan bool, c.NumThread)
	for i := 1; i <= max_i; i++ {
		sem <- true

		// Read chunk
		readBytes, err := input_file.Read(read_buffer)
		if err != nil {
			return err
		}

		// Hand over to thread to encode and write to file
		go func(i int, chunk string) {
			qr_image, err := encoder.Encode(chunk, gozxing.BarcodeFormat_QR_CODE, c.QRSideLength, c.QRSideLength, encode_hints)
			if err != nil {
				panic(err)
			}

			filename := c.FilePrefix + lpad(strconv.Itoa(i), "0", c.FilenamePadLength) + ".png"
			f, err := os.Create(output_folderpath + string(os.PathSeparator) + filename)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			if err := png.Encode(f, qr_image); err != nil {
				panic(err)
			}

			// Update progress
			tempProgress := int((i * 100) / max_i)
			if tempProgress > *progress {
				*progress = tempProgress
			}

			<-sem
		}(i, string(read_buffer[:readBytes]))
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
func (c *converter) Decode(input_folderpath, output_filepath string, progress *int) error {
	// Init decoder twice with different settings as fallback.
	// See https://github.com/makiuchi-d/gozxing/issues/14#issuecomment-486279433 for details.
	decoder := qrcode.NewQRCodeReader()
	decode_hints := make(map[gozxing.DecodeHintType]interface{})
	decode_hints[gozxing.DecodeHintType_PURE_BARCODE] = true
	decode_hints2 := make(map[gozxing.DecodeHintType]interface{})

	output_file, err := os.Create(output_filepath)
	if err != nil {
		return err
	}
	defer output_file.Close()

	// Read files from folder to check how many are available
	files, err := ioutil.ReadDir(input_folderpath)
	if err != nil {
		return err
	}
	lock := sync.Mutex{}

	i_start := 1
	i_end := len(files)
	result_assemble := make(map[int][]byte, i_end)
	sem := make(chan bool, c.NumThread+1)

	// To reduce memory, we start a thread to stream memory to disk.
	// It also needs to take a thread, so that the last files are written before the program exits.
	go func(current_i, end_i int, progress *int, output_file *os.File, result_assemble map[int][]byte, lock *sync.Mutex, sem chan bool) {
		sem <- true

		stream_to_file(i_start, i_end, progress, output_file, result_assemble, lock, sem)
	}(i_start, i_end, progress, output_file, result_assemble, &lock, sem)

	// Iterating over files
	for i := i_start; i <= i_end; i++ {
		sem <- true

		// Reading and decoding in threads
		go func(i int, result_assemble map[int][]byte, lock *sync.Mutex) {
			filename := c.FilePrefix + lpad(strconv.Itoa(i), "0", c.FilenamePadLength) + ".png"
			f, err := os.Open(input_folderpath + string(os.PathSeparator) + filename)
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

			decodeResult, err := decoder.Decode(bmp, decode_hints)
			if err != nil {
				decodeResult, err = decoder.Decode(bmp, decode_hints2)

				if err != nil {
					panic(err)
				}
			}

			lock.Lock()
			result_assemble[i] = []byte(decodeResult.GetText())
			lock.Unlock()

			<-sem
		}(i, result_assemble, &lock)
	}

	// Wait for all threads
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return nil
}

// stream_to_file writes content continuously from memory to disk to free memory
// while maintaining the order when writing to file
func stream_to_file(current_i, end_i int, progress *int, output_file *os.File, result_assemble map[int][]byte, lock *sync.Mutex, sem chan bool) {
	for i := current_i; i <= end_i; i++ {
		if chunk, ok := result_assemble[i]; ok {
			tempProgress := int((current_i * 100) / end_i)
			if tempProgress > *progress {
				*progress = tempProgress
			}

			// Write to disk
			output_file.Write(chunk)

			// Free memory
			lock.Lock()
			result_assemble[i] = nil
			delete(result_assemble, i)
			lock.Unlock()
		} else {
			// Not ready yet, sleep & restart
			time.Sleep(2 * time.Second)
			stream_to_file(i, end_i, progress, output_file, result_assemble, lock, sem)
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
