package fmp4

import (
	"flvFileReader"
	"fmt"
	"log"
)

type FMP4Slice struct {
	Data  []byte
	Idx   int  //-1 for init
	Video bool //audio or video
}

type FMP4Creater struct {
	videoIdx    int
	videoInited bool
	audioIdx    int
	audioInited bool
}

func (this *FMP4Creater) AddFlvTag(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {
	switch tag.TagType {
	case flvFileReader.FLV_TAG_ScriptData:
		return
	case flvFileReader.FLV_TAG_Audio:
		slice = this.handleAudioTag(tag)
		return
	case flvFileReader.FLV_TAG_Video:
		slice = this.handleVideoTag(tag)
		return
	default:
		return
	}
}

func (this *FMP4Creater) handleAudioTag(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {
	if this.audioInited == false {
		this.audioInited = true
		return this.createAudioInitSeg(tag)
	} else {
		return this.createAudioSeg(tag)
	}
	return
}

func (this *FMP4Creater) handleVideoTag(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {
	if tag.Data[0] != 0x17 && tag.Data[0] != 0x27 {
		log.Println(fmt.Sprintf("%d not support now", int(tag.Data[0])))
		return
	}
	pktType := tag.Data[1]
	CompositionTime := 0
	cur := 2
	if pktType == 1 {
		CompositionTime = ((int(tag.Data[cur+0])) << 16) | ((int(tag.Data[cur+1])) << 8) | ((int(tag.Data[cur+2])) << 0)
		log.Println(CompositionTime)
	}
	cur += 3
	if this.videoInited == false {
		if pktType != 0 {
			log.Println("AVC pkt not find")
			return
		}
		this.videoInited = true
		return this.createVideoInitSeg(tag)
	} else {
		//one tag,one slice
		//one tag,may not one frame
		return this.createVideoSeg(tag)
	}
	return
}

func (this *FMP4Creater) createVideoInitSeg(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {
	slice = &FMP4Slice{}
	slice.Video = true
	slice.Idx = -1
	segEncoder := flvFileReader.AMF0Encoder{}
	segEncoder.Init()
	return
}

func (this *FMP4Creater) createVideoSeg(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {
	log.Fatal("vvv1")
	return
}

func (this *FMP4Creater) createAudioInitSeg(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {
	log.Fatal("aaa0")
	return
}

func (this *FMP4Creater) createAudioSeg(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {
	log.Fatal("aaa1")
	return
}
