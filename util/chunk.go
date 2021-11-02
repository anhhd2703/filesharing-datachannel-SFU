package util

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

func Detach(dc *webrtc.DataChannel) {
	defer elapsed("Detach")()
	// fileToBeChunked := "./demo.rar" // change here!
	fileToBeChunked := "../music.mp3" // change here!
	file, err := os.Open(fileToBeChunked)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 1*(1<<15) + 1*(1<<11) // 1 MB, change this to your requirement
	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)
	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)
		file.Read(partBuffer)
		fileName := "bigfile_" + strconv.FormatUint(i, 10)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// data := &proto.Data{
		// 	Id:   fileName,
		// 	Buff: partBuffer,
		// }
		// d, _ := json.Marshal(data)
		// dc.Send(d)
		WriteToLocal(partBuffer, fileName)
		// fmt.Println("Split to : ", partSize)
	}
	dc.SendText(fmt.Sprintf("%v", totalPartsNum))
	file.Close()
}

func Combine(totalPartsNum uint64) {
	newFileName := "NEWbigfile.mp4"
	_, err := os.Create(newFileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//set the newFileName file to APPEND MODE!!
	// open files r and w

	file, err := os.OpenFile(newFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// IMPORTANT! do not defer a file.Close when opening a file for APPEND mode!
	// defer file.Close()

	// just information on which part of the new file we are appending
	var writePosition int64 = 0

	for j := uint64(0); j < totalPartsNum; j++ {

		//read a chunk
		currentChunkFileName := "bigfile_" + strconv.FormatUint(j, 10)

		newFileChunk, err := os.Open(currentChunkFileName)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer newFileChunk.Close()

		chunkInfo, err := newFileChunk.Stat()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// calculate the bytes size of each chunk
		// we are not going to rely on previous data and constant

		var chunkSize int64 = chunkInfo.Size()
		chunkBufferBytes := make([]byte, chunkSize)

		fmt.Println("Appending at position : [", writePosition, "] bytes")
		writePosition = writePosition + chunkSize

		// read into chunkBufferBytes
		reader := bufio.NewReader(newFileChunk)
		_, err = reader.Read(chunkBufferBytes)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// DON't USE ioutil.WriteFile -- it will overwrite the previous bytes!
		// write/save buffer to disk
		//ioutil.WriteFile(newFileName, chunkBufferBytes, os.ModeAppend)

		_, err = file.Write(chunkBufferBytes)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		file.Sync() //flush to disk

		// free up the buffer for next cycle
		// should not be a problem if the chunk size is small, but
		// can be resource hogging if the chunk size is huge.
		// also a good practice to clean up your own plate after eating

		chunkBufferBytes = nil // reset or empty our buffer
	}

	// now, we close the newFileName
	file.Close()
}

func WriteToLocal(partBuffer []byte, fileName string) {
	// var data proto.Data
	// json.Unmarshal(partBuffer, &data)
	// fileName := data.Id
	// fmt.Println("-------------WriteToLocal------------------=%v", fileName)
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ioutil.WriteFile(fileName, partBuffer, os.ModeAppend)
	f.Sync()
	f.Close()
}

func WriteToFile(partBuffer []byte) {
	newFileName := "WriteToFile.mp4"
	_, err := os.Create(newFileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//set the newFileName file to APPEND MODE!!
	// open files r and w

	file, err := os.OpenFile(newFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	n, err := file.Write(partBuffer)
	fmt.Println("Written ", n, " bytes")
	file.Sync()
	file.Close()
}

func DetachGoRoutine() {
	defer elapsed("DetachGoRoutine")()
	var wg sync.WaitGroup
	fileToBeChunked := "../music.mp3"
	file, err := os.Open(fileToBeChunked)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 1*(1<<15) + 1*(1<<11)
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)
	for j := uint64(0); j < totalPartsNum; j++ {
		wg.Add(1)
		go func(j uint64) {
			defer wg.Done()
			buffer := make([]byte, fileChunk)
			_, err := file.ReadAt(buffer, int64(j*fileChunk))
			if err != nil {
				fmt.Println(err)
				return
			}
			fileName := "bigfile_1" + strconv.FormatUint(j, 10)
			WriteToLocal(buffer, fileName)
		}(j)
	}
	wg.Wait()
}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}

func main() {
	defer elapsed("page")() // <-- The trailing () is the deferred call
	time.Sleep(time.Second * 2)
}
