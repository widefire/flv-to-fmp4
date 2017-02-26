package fmp4

import (
	"aac"
	"flvFileReader"
	"fmt"
	"log"
	"os"
)

const (
	video_trak = 1
	audio_trak = 2
)
const (
	PCM_Platform_endian         = 0
	ADPCM                       = 1
	MP3                         = 2
	PCM_little_endian           = 3
	Nellymoser_16_mono          = 4
	Nellymoser_8_mono           = 5
	Nellymoser                  = 6
	G711_A_law_logarithmic_PCM  = 7
	G711_mu_law_logarithmic_PCM = 8
	AAC                         = 10
	Speex                       = 11
	MP3_8                       = 14
	Device_specific_sound       = 15
)

type FMP4Slice struct {
	Data  []byte
	Idx   int  //1 base,0 for init
	Video bool //audio or video
}

type FMP4Creater struct {
	videoIdx    int
	videoInited bool
	audioIdx    int
	audioInited bool

	width           int
	height          int
	fps             int
	audioSampleSize uint32
	audioSampleRate uint32
	audioType       int
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
	slice.Idx = 0
	this.videoIdx++
	segEncoder := flvFileReader.AMF0Encoder{}
	segEncoder.Init()
	//ftyp
	ftyp := &MP4Box{}
	ftyp.Push([]byte("ftyp"))
	ftyp.PushBytes([]byte("isom"))
	ftyp.Push4Bytes(0x200)
	ftyp.PushBytes([]byte("isom"))
	ftyp.PushBytes([]byte("avc1"))
	ftyp.Pop()
	err := segEncoder.AppendByteArray(ftyp.Flush())
	if err != nil {
		log.Println(err.Error())
		return
	}
	//moov
	moovBox := &MP4Box{}
	moovBox.Push([]byte("moov"))
	//mvhd
	moovBox.Push([]byte("mvhd"))
	moovBox.Push4Bytes(0)          //version
	moovBox.Push4Bytes(0)          //creation_time
	moovBox.Push4Bytes(0)          //modification_time
	moovBox.Push4Bytes(1000)       //time_scale
	moovBox.Push4Bytes(0xffffffff) //duration 1s
	log.Println("duration 0xffffffff now")
	moovBox.Push4Bytes(0x00010000) //rate
	moovBox.Push2Bytes(0x0100)     //volume
	moovBox.Push2Bytes(0)          //reserved
	moovBox.Push8Bytes(0)          //reserved
	moovBox.Push4Bytes(0x00010000) //matrix
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0) //matrix
	moovBox.Push4Bytes(0x00010000)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0) //matrix
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x40000000)
	moovBox.Push4Bytes(0x0) //pre_defined
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	//nextrack id
	moovBox.Push4Bytes(0xffffffff)
	//!mvhd
	moovBox.Pop()
	//trak
	moovBox.Push([]byte("trak"))
	//tkhd
	moovBox.Push([]byte("tkhd"))
	moovBox.Push4Bytes(0x07) //version and flag
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(video_trak) //track id
	moovBox.Push4Bytes(0)          //reserved
	moovBox.Push4Bytes(0xffffffff) //duration
	log.Println("duration 0xffffffff")
	moovBox.Push8Bytes(0)          //reserved
	moovBox.Push2Bytes(0)          //layer
	moovBox.Push2Bytes(0)          //alternate_group
	moovBox.Push2Bytes(0)          //volume
	moovBox.Push2Bytes(0)          //reserved
	moovBox.Push4Bytes(0x00010000) //matrix
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0) //matrix
	moovBox.Push4Bytes(0x00010000)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0) //matrix
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x40000000) //matrix
	//parse sps ,get w h fps
	this.width, this.height, this.fps = flvFileReader.ParseSPS(tag.Data[13:])
	moovBox.Push4Bytes(uint32(this.width << 16))  //width
	moovBox.Push4Bytes(uint32(this.height << 16)) //height
	//!tkhd
	moovBox.Pop()
	//mdia
	moovBox.Push([]byte("mdia"))
	//mdhd
	moovBox.Push([]byte("mdhd"))
	moovBox.Push4Bytes(0)          //version and flag
	moovBox.Push4Bytes(0)          //creation_time
	moovBox.Push4Bytes(0)          //modification_time
	moovBox.Push4Bytes(1000)       //time scale
	moovBox.Push4Bytes(0xffffffff) //duration
	log.Println("duration 0xffffffff")
	moovBox.Push4Bytes(0x55c40000) //language und
	//!mdhd
	moovBox.Pop()
	//hdlr
	moovBox.Push([]byte("hdlr"))
	moovBox.Push4Bytes(0) //version and flag
	moovBox.Push4Bytes(0) //reserved
	moovBox.PushBytes([]byte("vide"))
	moovBox.Push4Bytes(0) //reserved
	moovBox.Push4Bytes(0) //reserved
	moovBox.Push4Bytes(0) //reserved
	moovBox.PushBytes([]byte("VideoHandler"))
	moovBox.PushByte(0)
	//!hdlr
	moovBox.Pop()
	//minf
	moovBox.Push([]byte("minf"))
	//vmhd
	moovBox.Push([]byte("vmhd"))
	moovBox.Push4Bytes(1) //
	moovBox.Push2Bytes(0) //copy
	moovBox.Push2Bytes(0) //opcolor
	moovBox.Push2Bytes(0) //opcolor
	moovBox.Push2Bytes(0) //opcolor
	//dinf
	moovBox.Push([]byte("dinf"))
	//dref
	moovBox.Push([]byte("dref"))
	moovBox.Push4Bytes(0) //version
	moovBox.Push4Bytes(1) //entry_count
	//url
	moovBox.Push([]byte("url "))
	moovBox.Push4Bytes(1)
	//!url
	moovBox.Pop()
	//!dref
	moovBox.Pop()
	//!dinf
	moovBox.Pop()
	//stbl
	moovBox.Push([]byte("stbl"))
	this.stsdV(moovBox, tag) //stsd
	//stts
	moovBox.Push([]byte("stts"))
	moovBox.Push4Bytes(0) //version
	moovBox.Push4Bytes(0) //count
	//!stts
	moovBox.Pop()
	//stsc
	moovBox.Push([]byte("stsc"))
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	//!stsc
	moovBox.Pop()
	//stsz
	moovBox.Push([]byte("stsz"))
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	//!stsz
	moovBox.Pop()
	//stco
	moovBox.Push([]byte("stco"))
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	//!stco
	moovBox.Pop()
	//!stbl
	moovBox.Pop()
	//!vmhd
	moovBox.Pop()
	//!minf
	moovBox.Pop()
	//!mdia
	moovBox.Pop()
	//!trak
	moovBox.Pop()
	//mvex
	moovBox.Push([]byte("mvex"))
	//trex
	moovBox.Push([]byte("trex"))
	moovBox.Push4Bytes(0)          //version and flag
	moovBox.Push4Bytes(video_trak) //track id
	moovBox.Push4Bytes(1)          //
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0x00010001)
	//!trex
	moovBox.Pop()
	//!mvex
	moovBox.Pop()
	//!moov
	moovBox.Pop()

	err = segEncoder.AppendByteArray(moovBox.Flush())
	if err != nil {
		log.Println(err.Error())
		return
	}
	slice.Data, err = segEncoder.GetData()
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(slice)
	fp, err := os.OpenFile("initV.mp4", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer fp.Close()
	fp.Write(slice.Data)
	return
}

func (this *FMP4Creater) createVideoSeg(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {
	slice = &FMP4Slice{}
	slice.Video = true
	slice.Idx = this.videoIdx
	this.videoIdx++
	segEncoder := flvFileReader.AMF0Encoder{}
	segEncoder.Init()

	videBox := &MP4Box{}
	//moof
	videBox.Push([]byte("moof"))
	//mfhd
	videBox.Push([]byte("mfhd"))
	videBox.Push4Bytes(0) //version and flags
	videBox.Push4Bytes(uint32(slice.Idx))
	//mfhd
	videBox.Pop()
	//traf
	videBox.Push([]byte("traf"))
	//tfhd
	videBox.Push([]byte("tfhd"))
	videBox.Push4Bytes(0)          //version and flags
	videBox.Push4Bytes(video_trak) //track
	//!tfhd
	videBox.Pop()
	//tfdt
	videBox.Push([]byte("tfdt"))
	videBox.Push4Bytes(0)
	videBox.Push4Bytes(tag.Timestamp)
	//!tfdt
	videBox.Pop()
	//trun
	videBox.Push([]byte("trun"))
	videBox.Push4Bytes(0xf01) //0x01:data off set; 0x100|0x200|0x400|0x800=0xf
	videBox.Push4Bytes(1)     //1 sample
	videBox.Push4Bytes(0x79)  //data offset
	videBox.Push4Bytes(0x21)//sample flags
	trun 和 MP3未完成 先弄MP3
	//!trun
	videBox.Pop()
	//!traf
	videBox.Pop()
	//!moof
	videBox.Pop()
	return
}

func (this *FMP4Creater) createAudioInitSeg(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {

	this.audioType = int(tag.Data[0] >> 4)
	switch this.audioType {
	case MP3:
		this.audioSampleSize = 1152
		log.Fatal("mp3 audiso sample size not processed")
	case AAC:
		this.audioSampleSize = 1024
		asc := aac.GenerateAudioSpecificConfig(tag.Data[2:])
		this.audioSampleRate = uint32(asc.SamplingFrequency)
	default:
		log.Fatal("unknown audio type")
	}

	slice = &FMP4Slice{}
	slice.Video = true
	slice.Idx = 0
	this.audioIdx++
	segEncoder := flvFileReader.AMF0Encoder{}
	segEncoder.Init()
	//ftyp
	ftyp := &MP4Box{}
	ftyp.Push([]byte("ftyp"))
	ftyp.PushBytes([]byte("isom"))
	ftyp.Push4Bytes(1)
	ftyp.PushBytes([]byte("isom"))
	ftyp.PushBytes([]byte("avc1"))
	ftyp.Pop()
	err := segEncoder.AppendByteArray(ftyp.Flush())
	if err != nil {
		log.Println(err.Error())
		return
	}
	//moov
	moovBox := &MP4Box{}
	moovBox.Push([]byte("moov"))
	//mvhd
	moovBox.Push([]byte("mvhd"))
	moovBox.Push4Bytes(0)          //version
	moovBox.Push4Bytes(0)          //creation_time
	moovBox.Push4Bytes(0)          //modification_time
	moovBox.Push4Bytes(1000)       //time_scale
	moovBox.Push4Bytes(0xffffffff) //duration 1s
	log.Println("duration 0xffffffff now")
	moovBox.Push4Bytes(0x00010000) //rate
	moovBox.Push2Bytes(0x0100)     //volume
	moovBox.Push2Bytes(0)          //reserved
	moovBox.Push8Bytes(0)          //reserved
	moovBox.Push4Bytes(0x00010000) //matrix
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0) //matrix
	moovBox.Push4Bytes(0x00010000)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0) //matrix
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x40000000)
	moovBox.Push4Bytes(0x0) //pre_defined
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	//nextrack id
	moovBox.Push4Bytes(0xffffffff)
	//!mvhd
	moovBox.Pop()
	//trak
	moovBox.Push([]byte("trak"))
	//tkhd
	moovBox.Push([]byte("tkhd"))
	moovBox.Push4Bytes(0x07) //version and flag
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(audio_trak) //track id
	moovBox.Push4Bytes(0)          //reserved
	moovBox.Push4Bytes(0xffffffff) //duration
	log.Println("duration 0xffffffff")
	moovBox.Push8Bytes(0)          //reserved
	moovBox.Push2Bytes(0)          //layer
	moovBox.Push2Bytes(0)          //alternate_group
	moovBox.Push2Bytes(0x0100)     //volume
	moovBox.Push2Bytes(0)          //reserved
	moovBox.Push4Bytes(0x00010000) //matrix
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0) //matrix
	moovBox.Push4Bytes(0x00010000)
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x0) //matrix
	moovBox.Push4Bytes(0x0)
	moovBox.Push4Bytes(0x40000000) //matrix
	moovBox.Push4Bytes(0)          //width
	moovBox.Push4Bytes(0)          //height
	//!tkhd
	moovBox.Pop()
	//mdia
	moovBox.Push([]byte("mdia"))
	//mdhd
	moovBox.Push([]byte("mdhd"))
	moovBox.Push4Bytes(0) //version and flag
	moovBox.Push4Bytes(0) //creation_time
	moovBox.Push4Bytes(0) //modification_time
	log.Println("maybe to audio sample hz")
	moovBox.Push4Bytes(1000)       //time scale
	moovBox.Push4Bytes(0xffffffff) //duration
	log.Println("duration 0xffffffff")
	moovBox.Push4Bytes(0x55c40000) //language und
	//!mdhd
	moovBox.Pop()
	//hdlr
	moovBox.Push([]byte("hdlr"))
	moovBox.Push4Bytes(0) //version and flag
	moovBox.Push4Bytes(0) //reserved
	moovBox.PushBytes([]byte("soun"))
	moovBox.Push4Bytes(0) //reserved
	moovBox.Push4Bytes(0) //reserved
	moovBox.Push4Bytes(0) //reserved
	moovBox.PushBytes([]byte("AudioHandler"))
	moovBox.PushByte(0)
	//!hdlr
	moovBox.Pop()
	//minf
	moovBox.Push([]byte("minf"))
	//smhd
	moovBox.Push([]byte("smhd"))
	moovBox.Push4Bytes(0) //version and flag
	moovBox.Push2Bytes(0) //balance
	moovBox.Push2Bytes(0) //reserved
	//dinf
	moovBox.Push([]byte("dinf"))
	//dref
	moovBox.Push([]byte("dref"))
	moovBox.Push4Bytes(0) //version
	moovBox.Push4Bytes(1) //entry_count
	//url
	moovBox.Push([]byte("url "))
	moovBox.Push4Bytes(1)
	//!url
	moovBox.Pop()
	//!dref
	moovBox.Pop()
	//!dinf
	moovBox.Pop()
	//stbl
	moovBox.Push([]byte("stbl"))
	this.stsdA(moovBox, tag) //stsd
	//stts
	moovBox.Push([]byte("stts"))
	moovBox.Push4Bytes(0) //version
	moovBox.Push4Bytes(0) //count
	//!stts
	moovBox.Pop()
	//stsc
	moovBox.Push([]byte("stsc"))
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	//!stsc
	moovBox.Pop()
	//stsz
	moovBox.Push([]byte("stsz"))
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	//!stsz
	moovBox.Pop()
	//stco
	moovBox.Push([]byte("stco"))
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	//!stco
	moovBox.Pop()
	//!stbl
	moovBox.Pop()
	//!smhd
	moovBox.Pop()
	//!minf
	moovBox.Pop()
	//!mdia
	moovBox.Pop()
	//!trak
	moovBox.Pop()
	//mvex
	moovBox.Push([]byte("mvex"))
	//trex
	moovBox.Push([]byte("trex"))
	moovBox.Push4Bytes(0)          //version and flag
	moovBox.Push4Bytes(audio_trak) //track id
	moovBox.Push4Bytes(1)          //
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0)
	moovBox.Push4Bytes(0x00010001)
	//!trex
	moovBox.Pop()
	//!mvex
	moovBox.Pop()
	//!moov
	moovBox.Pop()

	err = segEncoder.AppendByteArray(moovBox.Flush())
	if err != nil {
		log.Println(err.Error())
		return
	}
	slice.Data, err = segEncoder.GetData()
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(slice)
	fp, err := os.Create("initA.mp4")
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer fp.Close()
	fp.Write(slice.Data)
	return
}

func (this *FMP4Creater) createAudioSeg(tag *flvFileReader.FlvTag) (slice *FMP4Slice) {
	log.Fatal("aaa1")
	return
}

func (this *FMP4Creater) stsdV(box *MP4Box, tag *flvFileReader.FlvTag) {
	//stsd
	box.Push([]byte("stsd"))
	box.Push4Bytes(0)
	box.Push4Bytes(1)
	//avc1
	box.Push([]byte("avc1"))
	box.Push4Bytes(0)
	box.Push2Bytes(0)
	box.Push2Bytes(1)
	box.Push2Bytes(0)
	box.Push2Bytes(0)
	box.Push4Bytes(0)
	box.Push4Bytes(0)
	box.Push4Bytes(0)
	box.Push2Bytes(uint16(this.width))
	box.Push2Bytes(uint16(this.height))
	box.Push4Bytes(0x00480000)
	box.Push4Bytes(0x00480000)
	box.Push4Bytes(0)
	box.Push2Bytes(1)
	box.PushByte(uint8(len("fmp4 coding")))
	box.PushBytes([]byte("fmp4 coding"))
	spaceEnd := make([]byte, 32-1-len("fmp4 coding"))
	box.PushBytes(spaceEnd)
	box.Push2Bytes(0x18)
	box.Push2Bytes(0xffff)
	//avcC
	box.Push([]byte("avcC"))
	box.PushBytes(tag.Data[5:])
	//!avcC
	box.Pop()
	//!avc1
	box.Pop()
	//!stsd
	box.Pop()
	return
}

func (this *FMP4Creater) stsdA(box *MP4Box, tag *flvFileReader.FlvTag) {
	//stsd
	box.Push([]byte("stsd"))
	box.Push4Bytes(0)
	box.Push4Bytes(1)
	//mp4a
	box.Push([]byte("mp4a"))
	box.Push4Bytes(0)                          //reserved
	box.Push2Bytes(0)                          //reserved
	box.Push2Bytes(1)                          //data reference index
	box.Push8Bytes(0)                          //reserved int32[2]
	box.Push2Bytes(2)                          //channel count
	box.Push2Bytes(16)                         //sample size
	box.Push2Bytes(0)                          //pre defined
	box.Push2Bytes(0)                          //reserved
	box.Push4Bytes(this.audioSampleRate << 16) //samplerate
	//esds
	box.Push([]byte("esds"))
	box.Push4Bytes(0) //version and flag
	box.PushByte(3)   //tag
	esd := &MP4Box{}
	esd.Push2Bytes(0) //ES ID
	esd.PushByte(0)   //1:streamDependenceFlag=0  1:URL_Flag=0 1:OCRstreamFlag=0 5:streamPrority=0
	esd.PushByte(4)   //DecoderConfigDescriptor tag
	esdDesc := &MP4Box{}
	switch this.audioType { //object type indication
	case MP3:
		esdDesc.PushByte(0x6b)
	case AAC:
		esdDesc.PushByte(0x40)
	default:
		esdDesc.PushByte(0x40)
		log.Println(fmt.Sprintf("audio type %d not support", this.audioType))
	}
	esdDesc.PushByte(0x15)      //固定15  streamType upstream reserved
	esdDesc.PushByte(0)         //24位buffer size db
	esdDesc.Push2Bytes(0x600)   //24位补充
	esdDesc.Push4Bytes(0x1f400) //max bitrate
	esdDesc.Push4Bytes(0x1f400) //avg bitrate
	if this.audioType == AAC {
		esdDesc.PushByte(0x05)
		if len(tag.Data) >= 2 {
			esdDesc.PushByte(byte(len(tag.Data) - 2))
			esdDesc.PushBytes(tag.Data[2:])
		}
	}
	esdDescData := esdDesc.Flush()
	esd.PushByte(byte(len(esdDescData)))
	esd.PushBytes(esdDescData)
	esd.PushByte(0x06) //SLConfigDescrTag
	esd.PushByte(0x01) //length field
	esd.PushByte(0x02) //predefined 0x02 reserved for use int mp4 faile
	esdData := esd.Flush()
	box.PushByte(byte(len(esdData)))
	box.PushBytes(esdData)
	//!esds
	box.Pop()
	//!mp4a
	box.Pop()
	//!stsd
	box.Pop()
	return
}
