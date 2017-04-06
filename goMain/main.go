package main

import (
	"flvFileReader"
	"fmp4"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	log.Println("aaa")
	os.Mkdir("video", 0777)
	os.Mkdir("audio", 0777)
	flvSrc := flvFileReader.FlvFileReader{}
	flvSrc.Init("sample.flv")
	fmp4Sinker := fmp4.FMP4Creater{}
	defer flvSrc.Close()
	times := 0
	for {
		tag := flvSrc.GetNextTag()
		if tag == nil {

			return
		}
		//		log.Println(tag.Timestamp)
		fmp4Sinker.AddFlvTag(tag)
		if times > 200 {
			return
		}
		times++
	}
}
