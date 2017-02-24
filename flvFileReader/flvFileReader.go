package flvFileReader

import (
	"bytes"
	"errors"
	"log"
	"os"
)

const (
	FLV_TAG_Audio      = 8
	FLV_TAG_Video      = 9
	FLV_TAG_ScriptData = 18
)

type FlvTag struct {
	TagType   uint8
	Timestamp uint32
	StreamID  uint32
	Data      []byte
}

type FlvFileReader struct {
	fp *os.File
}

func (this *FlvFileReader) Init(fileName string) (err error) {
	if this.fp != nil {
		this.fp.Close()
		this.fp = nil
	}
	this.fp, err = os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		log.Println(err.Error())
		return
	}
	headerData, err := this.read(9)
	if err != nil {
		log.Println("read flv header failed")
		return
	}

	tmpFLV := []byte{'F', 'L', 'V', 1}
	if false == bytes.Equal(tmpFLV, headerData[0:4]) || (headerData[4]&0xfa) != 0 {
		log.Println("not flv")
		return
	}

	return
}

func (this *FlvFileReader) GetNextTag() (tag *FlvTag) {
	//read prevtag size
	_, err := this.read(4)
	if err != nil {
		log.Println("read prevtag size failed")
		return
	}
	tagHeaderSize := 11
	tagHeader, err := this.read(tagHeaderSize)
	if err != nil {
		log.Println("read flv tag failed")
		return
	}
	tag = &FlvTag{}
	cur := 0
	tag.TagType = tagHeader[cur]
	cur++
	tagDataSize := ((uint32(tagHeader[cur+0])) << 16) | ((uint32(tagHeader[cur+1])) << 8) | ((uint32(tagHeader[cur+2])) << 0)
	cur += 3
	tag.Timestamp = ((uint32(tagHeader[cur+0])) << 16) | ((uint32(tagHeader[cur+1])) << 8) | ((uint32(tagHeader[cur+2])) << 0) | ((uint32(tagHeader[cur+3])) << 4)
	cur += 4
	tag.StreamID = ((uint32(tagHeader[cur+0])) << 16) | ((uint32(tagHeader[cur+1])) << 8) | ((uint32(tagHeader[cur+2])) << 0)
	cur += 3
	tag.Data, err = this.read(int(tagDataSize))
	if err != nil {
		log.Println("read flv tag data failed")
		tag = nil
		return
	}
	return
}

func (this *FlvFileReader) Close() {
	if this.fp != nil {
		this.fp.Close()
		this.fp = nil
	}
}

func (this *FlvFileReader) read(n int) (data []byte, err error) {
	if this.fp == nil {
		log.Fatal("file not opened")
	}
	fileData := make([]byte, n)
	ret, err := this.fp.Read(fileData)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if ret != n {
		err = errors.New("no more data to read")
		log.Println(err.Error())
		return
	}
	data = fileData
	return
}
