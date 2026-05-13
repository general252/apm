#ifndef LK_APM_BRIDGE_H
#define LK_APM_BRIDGE_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

    typedef void* ApmHandle;

    /**
    * 会议
config.gain_controller2.enabled = true;
config.gain_controller2.input_volume_controller.enabled = false;
config.gain_controller2.adaptive_digital.enabled = true;
config.gain_controller2.adaptive_digital.headroom_db = 6.f;
config.gain_controller2.adaptive_digital.max_gain_db = 25.f;
config.gain_controller2.adaptive_digital.initial_gain_db = 8.f;
config.gain_controller2.adaptive_digital.max_gain_change_db_per_second = 4.f;
config.gain_controller2.adaptive_digital.max_output_noise_level_dbfs = -50.f;
config.gain_controller2.fixed_digital.gain_db = 4.f;
    */
    typedef struct {
        int32_t sample_rate_hz; // = 48000
        int32_t num_channels; // - 1


        bool gain_controller2_enabled;// = false;
        bool gain_controller2_adaptive_digital_enabled;// = false; // 自适应数字增益
        float gain_controller2_adaptive_digital_headroom_db;// = 5.0f;
        float gain_controller2_adaptive_digital_max_gain_db;// = 50.0f;
        float gain_controller2_adaptive_digital_initial_gain_db;// = 15.0f;
        float gain_controller2_adaptive_digital_max_gain_change_db_per_second;// = 6.0f;
        float gain_controller2_adaptive_digital_max_output_noise_level_dbfs;// = -50.0f;

        float gain_controller2_fixed_digital_gain_db;// = 0.0f; 固定数字增益（dB）10^(10/20)=3.16 3dB轻微增强 6dB约2倍 10dB明显增强(3.16)

        bool high_pass_filter_enabled;//= false; // 高通滤波, 去除：DC 低频震动 风噪 电流声 通常截止：80Hz
        bool high_pass_filter_apply_in_full_band;// = true;

        bool echo_canceller_enabled;// = false;
        bool echo_canceller_mobile_mode;// = false;
        bool echo_canceller_export_linear_aec_output;// = false;
        bool echo_canceller_enforce_high_pass_filtering;// = true;

        bool noise_suppression_enabled;// = false;
        // enum Level { kLow, kModerate, kHigh, kVeryHigh };
        int8_t noise_suppression_level; // = kModerate;
        bool noise_suppression_analyze_linear_aec_output_when_available;// = false;

        bool transient_suppression_enabled;// = false;

    } APMConfig;

    // Create an APM instance. Returns NULL on error, sets *err to non-zero.
    ApmHandle apm_create(APMConfig* param, int* err);

    // Destroy an APM instance.
    void apm_destroy(ApmHandle h);

    // Process a 10ms render (far-end/playback) frame in-place. Returns 0 on success.
    int apm_process_reverse_stream(ApmHandle h, float* samples, int num_channels);
    int apm_process_reverse_stream_int16(ApmHandle h, int16_t* samples, int num_channels);

    // Process a 10ms capture frame in-place. Returns 0 on success.
    int apm_process_stream(ApmHandle h, float* samples, int num_channels);
    int apm_process_stream_int16(ApmHandle h, int16_t* samples, int num_channels);

    // Set the stream delay in milliseconds for echo cancellation.
    void apm_set_stream_delay_ms(ApmHandle h, int delay_ms);

    // Get the current stream delay in milliseconds.
    int apm_stream_delay_ms(ApmHandle h);

    // AEC statistics returned by apm_get_stats.
    typedef struct {
        int    has_erl;
        double echo_return_loss;          // ERL in dB
        int    has_erle;
        double echo_return_loss_enhancement; // ERLE in dB
        int    has_divergent;
        double divergent_filter_fraction;
        int    has_delay;
        int    delay_ms;
        int    has_residual_echo;
        double residual_echo_likelihood;
    } ApmStats;

    // Get current AEC statistics.
    void apm_get_stats(ApmHandle h, ApmStats* out);

#ifdef __cplusplus
}
#endif

#endif // LK_APM_BRIDGE_H