package main

import (
	"flvFileReader"
	"fmp4"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
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
		//log.Println(tag.Timestamp)
		fmp4Sinker.AddFlvTag(tag)
		if times > 200 {
			return
		}
		times++
	}
}
