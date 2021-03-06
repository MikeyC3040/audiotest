package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"io"
	"sync"
	"unsafe"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

func init() {
	go C.gstreamer_send_start_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline   *C.GstElement
	audioTrack *webrtc.TrackLocalStaticSample
}

var pipeline = &Pipeline{}
var pipelinesLock sync.Mutex

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(audioTrack *webrtc.TrackLocalStaticSample) *Pipeline {
	pipelineStr := "audiotestsrc ! queue ! audioconvert ! audioresample ! opusenc ! appsink name=audio"

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()
	pipeline = &Pipeline{
		Pipeline:   C.gstreamer_send_create_pipeline(pipelineStrUnsafe),
		audioTrack: audioTrack,
	}
	return pipeline
}

// Start starts the GStreamer Pipeline
func (p *Pipeline) Start() {
	// This will signal to goHandlePipelineBuffer
	// and provide a method for cancelling sends.
	C.gstreamer_send_start_pipeline(p.Pipeline)
}

// Play sets the pipeline to PLAYING
func (p *Pipeline) Play() {
	C.gstreamer_send_play_pipeline(p.Pipeline)
}

// Pause sets the pipeline to PAUSED
func (p *Pipeline) Pause() {
	C.gstreamer_send_pause_pipeline(p.Pipeline)
}

// SeekToTime seeks on the pipeline
func (p *Pipeline) SeekToTime(seekPos int64) {
	C.gstreamer_send_seek(p.Pipeline, C.int64_t(seekPos))
}

const (
	audioClockRate = 48000
)

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(buffer unsafe.Pointer, bufferLen C.int, duration C.int, isVideo C.int) {
	var track *webrtc.TrackLocalStaticSample
	var samples uint32

	samples = uint32(audioClockRate * (float32(duration) / 1000000000))
	track = pipeline.audioTrack

	if err := track.WriteSample(media.Sample{Data: C.GoBytes(buffer, bufferLen), Samples: samples}); err != nil && err != io.ErrClosedPipe {
		panic(err)
	}

	C.free(buffer)
}
