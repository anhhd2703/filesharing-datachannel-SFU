package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
	"webrtc/proto"

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
	const fileChunk = 1*(1<<15) + 1*(1<<11)
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
		data := &proto.Data{
			Id:   fileName,
			Buff: partBuffer,
		}
		d, _ := json.Marshal(data)
		// WriteToLocal(data)
		dc.Send(d)
	}
	file.Close()
	dc.SendText(fmt.Sprintf("%v", totalPartsNum))
}

func Combine(totalPartsNum uint64) {
	defer elapsed("Combine")()
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

		n, err := file.Write(chunkBufferBytes)

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

		fmt.Println("Written ", n, " bytes")

		fmt.Println("Recombining part [", j, "] into : ", newFileName)
	}

	// now, we close the newFileName
	file.Close()
}

func WriteToLocal(data *proto.Data) {

	f, err := os.Create(data.Id)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ioutil.WriteFile(data.Id, data.Buff, os.ModeAppend)
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

func DetachGoRoutine(dc *webrtc.DataChannel) {
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
	poolSize := uint64(10)
	for j := uint64(0); j < poolSize; j++ {
		stepSize := uint64(math.Ceil(float64(totalPartsNum) / float64(poolSize)))
		wg.Add(1)
		go func(j uint64) {
			defer wg.Done()
			for k := uint64(j * stepSize); k < stepSize*(j+1); k++ {
				partSize := int(math.Min(fileChunk, float64(fileSize-int64(k*fileChunk))))
				if partSize < 0 {
					break
				}
				buffer := make([]byte, partSize)
				_, err := file.ReadAt(buffer, int64(k*fileChunk))
				if err != nil {
					fmt.Println(err)
					break
				}
				fileName := "bigfile_" + strconv.FormatUint(k, 10)
				data := &proto.Data{
					Id:   fileName,
					Buff: buffer,
				}
				d, _ := json.Marshal(data)
				dc.Send(d)
			}
		}(j)
	}
	wg.Wait()
	dc.SendText(fmt.Sprintf("%v", totalPartsNum))
}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}
