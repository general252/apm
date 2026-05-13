package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/general252/apm"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	_, _ = fmt.Fprintf(os.Stderr, "input option: ")
	line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		return
	}

	log.Printf("line: %v", string(line))

	switch string(line) {
	case "1":
		AEC3()
		log.Println("echo canceller")
	case "2":
		AGC2()
		log.Println("gain controller")
	default:
		log.Println("invalid option")
	}
}

func AGC2() {
	const SampleRateHz = 8000
	const NumChannels = 1
	const SampleFrameSize = SampleRateHz / 1000 * 10

	cfg := apm.NewConfig()
	cfg.SetSampleRateHz(SampleRateHz)
	cfg.SetNumChannels(NumChannels)
	cfg.SetGainController2(true)

	apm, err := apm.NewAPM(cfg)
	if err != nil {
		return
	}
	defer apm.Close()

	data1, _ := os.ReadFile("voice.pcm")
	data11 := BytesToInt16(data1)

	fout, _ := os.Create("voice_out.pcm")
	defer fout.Close()

	for i := 0; i < len(data11)/SampleFrameSize-1; i++ {
		dd1 := data11[i*SampleFrameSize : (i+1)*SampleFrameSize]

		//log.Println(dd2)
		err = apm.ProcessStreamInt16(dd1)
		if err != nil {
			log.Println(err)
		}
		//log.Println(dd2)

		//stats := apm.GetStats()
		//log.Printf("delay: %v, ERL: %v, ERLE: %v", stats.DelayMs, stats.EchoReturnLoss, stats.EchoReturnLossEnhancement)

		out := Int16ToBytes(dd1)
		_, _ = fout.Write(out)
	}
}

func AEC3() {
	const SampleRateHz = 8000
	const NumChannels = 1
	const SampleFrameSize = SampleRateHz / 1000 * 10

	cfg := apm.NewConfig()
	cfg.SetSampleRateHz(SampleRateHz)
	cfg.SetNumChannels(NumChannels)
	cfg.SetHighPassFilterEnabled(true)
	cfg.SetEchoCanceller(true)
	cfg.SetNoiseSuppressor(true)
	cfg.SetGainController2(false)

	apm, err := apm.NewAPM(cfg)
	if err != nil {
		return
	}
	defer apm.Close()

	apm.SetStreamDelayMs(500)

	data1, _ := os.ReadFile("01_far_end_2.pcm")
	data2, _ := os.ReadFile("02_mic_in_2.pcm")

	data11 := BytesToInt16(data1)
	data22 := BytesToInt16(data2)

	fout, _ := os.Create("03_out_2.pcm")
	defer fout.Close()

	for i := 0; i < len(data11)/SampleFrameSize-1; i++ {
		dd1 := data11[i*SampleFrameSize : (i+1)*SampleFrameSize]
		dd2 := data22[i*SampleFrameSize : (i+1)*SampleFrameSize]

		//log.Println(dd1)
		err = apm.ProcessReverseStreamInt16(dd1)
		if err != nil {
			log.Println(err)
		}

		//log.Println(dd2)
		err = apm.ProcessStreamInt16(dd2)
		if err != nil {
			log.Println(err)
		}
		//log.Println(dd2)

		//stats := apm.GetStats()
		//log.Printf("delay: %v, ERL: %v, ERLE: %v", stats.DelayMs, stats.EchoReturnLoss, stats.EchoReturnLossEnhancement)

		out := Int16ToBytes(dd2)
		_, _ = fout.Write(out)
	}
}

func bytesToInt16(data []byte) []int16 {
	const SampleSize = 2
	out := make([]int16, len(data)/SampleSize)
	for i := 0; i < len(out); i++ {
		out[i] = int16(binary.LittleEndian.Uint16(data[i*SampleSize : (i+1)*SampleSize]))
	}

	return out
}

func int16ToBytes(data []int16) []byte {
	const SampleSize = 2
	out := make([]byte, len(data)*SampleSize)
	for i := 0; i < len(data); i++ {
		binary.LittleEndian.PutUint16(out[i*SampleSize:(i+1)*SampleSize], uint16(data[i]))
	}
	return out
}

func BytesToInt16(b []byte) []int16 {
	i16 := unsafe.Slice((*int16)(unsafe.Pointer(&b[0])), len(b)/2)
	return i16
}

func Int16ToBytes(i16 []int16) []byte {
	b := unsafe.Slice((*byte)(unsafe.Pointer(&i16[0])), len(i16)*2)
	return b
}

func Float32SliceToBytes(fs []float32) []byte {
	if len(fs) == 0 {
		return nil
	}
	// 每个 float32 对应 4 个 byte
	return unsafe.Slice((*byte)(unsafe.Pointer(&fs[0])), len(fs)*4)
}

func ByteSliceToFloat32(b []byte) []float32 {
	if len(b) == 0 {
		return nil
	}
	// 长度需要除以 4
	return unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), len(b)/4)
}
