package apm

type NoiseSuppressorLevel = int8

const (
	kLow      NoiseSuppressorLevel = 0
	kModerate NoiseSuppressorLevel = 1
	kHigh     NoiseSuppressorLevel = 2
	kVeryHigh NoiseSuppressorLevel = 3
)

type APMConfig struct {
	sample_rate_hz int32 // = 48000
	num_channels   int32 // - 1

	high_pass_filter_enabled            bool // = false;
	high_pass_filter_apply_in_full_band bool // = true;

	echo_canceller_enabled                     bool // = false;
	echo_canceller_mobile_mode                 bool // = false;
	echo_canceller_export_linear_aec_output    bool // = false;
	echo_canceller_enforce_high_pass_filtering bool // = true;

	noise_suppression_enabled                                  bool                 // = false;
	noise_suppression_level                                    NoiseSuppressorLevel // = kModerate;
	noise_suppression_analyze_linear_aec_output_when_available bool                 // = false;

	gain_controller2_enabled                                        bool    // = false;
	gain_controller2_adaptive_digital_enabled                       bool    // = false; // 自适应数字增益
	gain_controller2_adaptive_digital_headroom_db                   float32 // = 5.0f;
	gain_controller2_adaptive_digital_max_gain_db                   float32 // = 50.0f;
	gain_controller2_adaptive_digital_initial_gain_db               float32 // = 15.0f;
	gain_controller2_adaptive_digital_max_gain_change_db_per_second float32 // = 6.0f;
	gain_controller2_adaptive_digital_max_output_noise_level_dbfs   float32 // = -50.0f;
	gain_controller2_fixed_digital_gain_db                          float32 // = 0.0f; 固定数字增益（dB）10^(10/20)=3.16 3dB轻微增强 6dB约2倍 10dB明显增强(3.16)

}

// NewConfig 创建一个新的 APMConfig 实例
/**
AudioRecord / ALSA / CoreAudio
        ↓
InputVolumeController（可选） 调系统麦克风音量
        ↓
CaptureLevelAdjustment.pre_gain
        ↓
PreAmplifier（旧）
        ↓
HighPassFilter
        ↓
AEC（EchoCanceller）
        ↓
NoiseSuppression
        ↓
TransientSuppression（旧）
        ↓
AGC2 AdaptiveDigital
        ↓
AGC2 FixedDigital
        ↓
Limiter
        ↓
CaptureLevelAdjustment.post_gain
        ↓
编码 Opus/AAC
*/
func NewConfig() *APMConfig {
	return &APMConfig{
		sample_rate_hz: 48000,
		num_channels:   1,

		high_pass_filter_enabled:            false,
		high_pass_filter_apply_in_full_band: true,

		echo_canceller_enabled:                     false,
		echo_canceller_mobile_mode:                 false,
		echo_canceller_export_linear_aec_output:    false,
		echo_canceller_enforce_high_pass_filtering: true,

		gain_controller2_enabled:                                        false, // = false;
		gain_controller2_adaptive_digital_enabled:                       true,  // = false; // 自适应数字增益
		gain_controller2_adaptive_digital_headroom_db:                   6.0,   // = 5.0f; // 标准会议 6;安静环境（桌面麦克风） 15;远场/手机免提 25~30;噪声环境（街道/车内）15~20
		gain_controller2_adaptive_digital_max_gain_db:                   20.0,  // = 50.0f;
		gain_controller2_adaptive_digital_initial_gain_db:               8.0,   // = 15.0f;
		gain_controller2_adaptive_digital_max_gain_change_db_per_second: 3.0,   // = 6.0f;
		gain_controller2_adaptive_digital_max_output_noise_level_dbfs:   -50.0, // = -50.0f;
		gain_controller2_fixed_digital_gain_db:                          0.0,   // = 0.0f; 固定数字增益（dB）10^(10/20)=3.16 3dB轻微增强 6dB约2倍 10dB明显增强(3.16)

		noise_suppression_enabled:                                  false,
		noise_suppression_level:                                    kHigh,
		noise_suppression_analyze_linear_aec_output_when_available: false,
	}
}

func (cfg *APMConfig) SetSampleRateHz(rate int32) {
	cfg.sample_rate_hz = rate
}

func (cfg *APMConfig) SetNumChannels(channels int32) {
	cfg.num_channels = channels
}

func (cfg *APMConfig) SetHighPassFilterEnabled(enabled bool) {
	cfg.high_pass_filter_enabled = enabled
}

func (cfg *APMConfig) SetEchoCanceller(enabled bool) {
	cfg.echo_canceller_enabled = enabled
}

func (cfg *APMConfig) SetNoiseSuppressor(enabled bool) {
	cfg.noise_suppression_enabled = enabled
}

func (cfg *APMConfig) SetGainController2(enabled bool) {
	cfg.gain_controller2_enabled = enabled
}

func (cfg *APMConfig) SetNoiseSuppressorLevel(level NoiseSuppressorLevel) {
	cfg.noise_suppression_level = level
}
