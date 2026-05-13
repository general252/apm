package apm

/*
#cgo CXXFLAGS: -std=c++17
#cgo LDFLAGS: -lstdc++
#cgo pkg-config: webrtc-audio-processing-2

#include "bridge.h"
*/
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

// 参考 https://github.com/livekit/livekit-cli/tree/main/pkg/apm

type APM struct {
	handle C.ApmHandle
	cfg    *APMConfig
}

func NewAPM(config *APMConfig) (*APM, error) {
	var cfg C.APMConfig

	cfg.sample_rate_hz = C.int32_t(config.sample_rate_hz)
	cfg.num_channels = C.int32_t(config.num_channels)

	cfg.echo_canceller_enabled = boolToInt(config.echo_canceller_enabled)
	cfg.echo_canceller_mobile_mode = boolToInt(config.echo_canceller_mobile_mode)
	cfg.echo_canceller_export_linear_aec_output = boolToInt(config.echo_canceller_export_linear_aec_output)
	cfg.echo_canceller_enforce_high_pass_filtering = boolToInt(config.echo_canceller_enforce_high_pass_filtering)

	cfg.gain_controller2_enabled = boolToInt(config.gain_controller2_enabled)
	cfg.gain_controller2_adaptive_digital_enabled = boolToInt(config.gain_controller2_adaptive_digital_enabled)
	cfg.gain_controller2_adaptive_digital_headroom_db = C.float(config.gain_controller2_adaptive_digital_headroom_db)
	cfg.gain_controller2_adaptive_digital_max_gain_db = C.float(config.gain_controller2_adaptive_digital_max_gain_db)
	cfg.gain_controller2_adaptive_digital_initial_gain_db = C.float(config.gain_controller2_adaptive_digital_initial_gain_db)
	cfg.gain_controller2_adaptive_digital_max_gain_change_db_per_second = C.float(config.gain_controller2_adaptive_digital_max_gain_change_db_per_second)
	cfg.gain_controller2_adaptive_digital_max_output_noise_level_dbfs = C.float(config.gain_controller2_adaptive_digital_max_output_noise_level_dbfs)
	cfg.gain_controller2_fixed_digital_gain_db = C.float(config.gain_controller2_fixed_digital_gain_db)

	cfg.high_pass_filter_enabled = boolToInt(config.high_pass_filter_enabled)
	cfg.high_pass_filter_apply_in_full_band = boolToInt(config.high_pass_filter_apply_in_full_band)

	cfg.noise_suppression_enabled = boolToInt(config.noise_suppression_enabled)
	cfg.noise_suppression_level = C.int8_t(config.noise_suppression_level)
	cfg.noise_suppression_analyze_linear_aec_output_when_available = boolToInt(config.noise_suppression_analyze_linear_aec_output_when_available) // = false;

	var cerr C.int
	handle := C.apm_create(
		&cfg,
		&cerr,
	)
	if handle == nil {
		return nil, fmt.Errorf("apm: failed to create audio processing module, cerr: %d", cerr)
	}

	if cerr != 0 {
		return nil, fmt.Errorf("apm: failed to create audio processing module, cerr: %d", cerr)
	}

	a := &APM{
		handle: handle,
		cfg:    config,
	}
	runtime.SetFinalizer(a, func(a *APM) { a.Close() })

	return a, nil
}

func (a *APM) Close() {
	if a.handle != nil {
		C.apm_destroy(a.handle)
		a.handle = nil
	}
}

func (a *APM) get10msCont() int {
	// 48000 / 1000 = 48 *10 = 480
	return int(a.cfg.sample_rate_hz / 100)
}

func (a *APM) ProcessReverseStream(samples []float32) error {
	if a.handle == nil {
		return fmt.Errorf("apm: closed")
	}
	if len(samples) == 0 {
		return nil
	}

	numChannels := a.cfg.num_channels

	ret := C.apm_process_reverse_stream(
		a.handle,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.int(numChannels),
	)
	if ret != 0 {
		return fmt.Errorf("apm: ProcessReverseStream failed, ret: %d", ret)
	}
	return nil
}

func (a *APM) ProcessReverseStreamInt16(samples []int16) error {
	if a.handle == nil {
		return fmt.Errorf("apm: closed")
	}
	if len(samples) == 0 {
		return nil
	}

	numChannels := a.cfg.num_channels

	ret := C.apm_process_reverse_stream_int16(
		a.handle,
		(*C.int16_t)(unsafe.Pointer(&samples[0])),
		C.int(numChannels),
	)
	if ret != 0 {
		return fmt.Errorf("apm: ProcessReverseStream failed, ret: %d", ret)
	}
	return nil
}

func (a *APM) ProcessStream(samples []float32) error {
	if a.handle == nil {
		return fmt.Errorf("apm: closed")
	}
	if len(samples) == 0 {
		return nil
	}

	numChannels := a.cfg.num_channels

	ret := C.apm_process_stream(
		a.handle,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.int(numChannels),
	)
	if ret != 0 {
		return fmt.Errorf("apm: ProcessStream failed, ret: %d", ret)
	}
	return nil
}

func (a *APM) ProcessStreamInt16(samples []int16) error {
	if a.handle == nil {
		return fmt.Errorf("apm: closed")
	}
	if len(samples) == 0 {
		return nil
	}

	numChannels := a.cfg.num_channels

	ret := C.apm_process_stream_int16(
		a.handle,
		(*C.int16_t)(unsafe.Pointer(&samples[0])),
		C.int(numChannels),
	)
	if ret != 0 {
		return fmt.Errorf("apm: ProcessStream failed, ret: %d", ret)
	}
	return nil
}

func (a *APM) SetStreamDelayMs(ms int) {
	if a.handle == nil {
		return
	}
	C.apm_set_stream_delay_ms(a.handle, C.int(ms))
}

func (a *APM) StreamDelayMs() int {
	if a.handle == nil {
		return 0
	}
	return int(C.apm_stream_delay_ms(a.handle))
}

// Stats holds AEC statistics from the WebRTC APM.
type Stats struct {
	EchoReturnLoss            float64 // ERL in dB (higher = more echo removed)
	EchoReturnLossEnhancement float64 // ERLE in dB (higher = better cancellation)
	DivergentFilterFraction   float64 // 0-1, fraction of time filter is divergent
	DelayMs                   int     // Estimated echo path delay
	ResidualEchoLikelihood    float64 // 0-1, likelihood of residual echo
	HasERL                    bool
	HasERLE                   bool
	HasDelay                  bool
	HasResidualEcho           bool
	HasDivergent              bool
}

// GetStats returns the current AEC statistics.
func (a *APM) GetStats() Stats {
	if a.handle == nil {
		return Stats{}
	}

	var cs C.ApmStats
	C.apm_get_stats(a.handle, &cs)

	return Stats{
		EchoReturnLoss:            float64(cs.echo_return_loss),
		EchoReturnLossEnhancement: float64(cs.echo_return_loss_enhancement),
		DivergentFilterFraction:   float64(cs.divergent_filter_fraction),
		DelayMs:                   int(cs.delay_ms),
		ResidualEchoLikelihood:    float64(cs.residual_echo_likelihood),
		HasERL:                    cs.has_erl != 0,
		HasERLE:                   cs.has_erle != 0,
		HasDelay:                  cs.has_delay != 0,
		HasResidualEcho:           cs.has_residual_echo != 0,
		HasDivergent:              cs.has_divergent != 0,
	}
}

func boolToInt(b bool) C.bool {
	if b {
		return C.bool(true)
	}
	return C.bool(false)
}
