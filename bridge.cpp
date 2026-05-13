#include "bridge.h"

#include "webrtc-audio-processing-2/api/audio/audio_processing.h"
#include <memory>

#include <stdio.h>
#include <stdint.h>
#include <inttypes.h> // 必须包含这个头文件

void print_samples(int16_t* samples, size_t count) {
    printf("Samples int16: [");
    for (size_t i = 0; i < count; i++) {
        // PRId16 会根据平台自动替换为 "d" 或 "hd"
        printf("%" PRId16 " ", samples[i]);
    }
    printf("]\n");
}

void print_formatted_floats(float* samples, size_t count) {
    fprintf(stderr, "Samples float: [");
    for (size_t i = 0; i < count; i++) {
        // 使用 stderr 打印，宽度 8，保留 4 位小数
        fprintf(stderr, "%8.4f ", samples[i]);
    }
    fprintf(stderr, "]\n");
}


struct ApmInstance {
    webrtc::scoped_refptr<webrtc::AudioProcessing> apm;
    int32_t sample_rate_hz;
    int32_t num_channels;
};


#ifdef __cplusplus
extern "C" {
#endif



    ApmHandle apm_create(APMConfig* param, int32_t* err)
    {
        rtc::scoped_refptr<webrtc::AudioProcessing> apm = webrtc::AudioProcessingBuilder().Create();
        if (!apm) {
            if (err) *err = -1;
            return nullptr;
        }

        webrtc::AudioProcessing::Config cfg;

        cfg.gain_controller1.enabled = false; // gain control 不使用低版本
        cfg.gain_controller2.enabled = param->gain_controller2_enabled; // gain control 增益增强
        cfg.gain_controller2.adaptive_digital.enabled = param->gain_controller2_adaptive_digital_enabled;// = false; // 自适应数字增益
        cfg.gain_controller2.adaptive_digital.headroom_db = param->gain_controller2_adaptive_digital_headroom_db;// = 5.0f;
        cfg.gain_controller2.adaptive_digital.max_gain_db = param->gain_controller2_adaptive_digital_max_gain_db;// = 50.0f;
        cfg.gain_controller2.adaptive_digital.initial_gain_db = param->gain_controller2_adaptive_digital_initial_gain_db;// = 15.0f;
        cfg.gain_controller2.adaptive_digital.max_gain_change_db_per_second = param->gain_controller2_adaptive_digital_max_gain_change_db_per_second;// = 6.0f;
        cfg.gain_controller2.adaptive_digital.max_output_noise_level_dbfs = param->gain_controller2_adaptive_digital_max_output_noise_level_dbfs;// = -50.0f;
        cfg.gain_controller2.fixed_digital.gain_db = param->gain_controller2_fixed_digital_gain_db; // 固定数字增益（dB）

        cfg.high_pass_filter.enabled = param->high_pass_filter_enabled; // High Pass Filter 高通滤波。滤除低频杂音（通常是 100Hz 以下），如手持设备的触摸声
        cfg.high_pass_filter.apply_in_full_band = param->high_pass_filter_apply_in_full_band;

        cfg.echo_canceller.enabled = param->echo_canceller_enabled; // echo cancellation 回声消除
        cfg.echo_canceller.mobile_mode = param->echo_canceller_mobile_mode;
        cfg.echo_canceller.export_linear_aec_output = param->echo_canceller_export_linear_aec_output;
        cfg.echo_canceller.enforce_high_pass_filtering = param->echo_canceller_enforce_high_pass_filtering;


        cfg.noise_suppression.enabled = param->noise_suppression_enabled; // noise suppression 降噪
        cfg.noise_suppression.level = (webrtc::AudioProcessing::Config::NoiseSuppression::Level)(param->noise_suppression_level); // 降噪级别
        cfg.noise_suppression.analyze_linear_aec_output_when_available = param->noise_suppression_analyze_linear_aec_output_when_available;

        apm->ApplyConfig(cfg);
        apm->Initialize();

        auto* inst = new ApmInstance{ std::move(apm), };
        inst->sample_rate_hz = param->sample_rate_hz;
        inst->num_channels = param->num_channels;

        if (err) *err = 0;
        return static_cast<ApmHandle>(inst);
    }

    void apm_destroy(ApmHandle h) {
        if (h) {
            delete static_cast<ApmInstance*>(h);
        }
    }


    void apm_set_stream_delay_ms(ApmHandle h, int32_t delay_ms) {
        auto* inst = static_cast<ApmInstance*>(h);
        inst->apm->set_stream_delay_ms(delay_ms);
    }

    int32_t apm_stream_delay_ms(ApmHandle h) {
        auto* inst = static_cast<ApmInstance*>(h);
        return inst->apm->stream_delay_ms();
    }


    int32_t apm_process_reverse_stream(ApmHandle h, float* samples, int32_t num_channels) {
        auto* inst = static_cast<ApmInstance*>(h);
        webrtc::StreamConfig stream_cfg(inst->sample_rate_hz, num_channels);

        float* channel_ptrs[1] = { samples };
        return inst->apm->ProcessReverseStream(channel_ptrs, stream_cfg, stream_cfg, channel_ptrs);
    }

    int32_t apm_process_reverse_stream_int16(ApmHandle h, int16_t* samples, int32_t num_channels)
    {
        auto* inst = static_cast<ApmInstance*>(h);
        webrtc::StreamConfig stream_cfg(inst->sample_rate_hz, num_channels);

        return inst->apm->ProcessReverseStream((const int16_t* const)samples, stream_cfg, stream_cfg, (int16_t* const)samples);
    }

    int32_t apm_process_stream(ApmHandle h, float* samples, int32_t num_channels) {
        // 10ms at 48kHz = 480 samples per channel

        auto* inst = static_cast<ApmInstance*>(h);
        webrtc::StreamConfig stream_cfg(inst->sample_rate_hz, num_channels);

        //print_formatted_floats(samples, 80);
        //fprintf(stderr, "inst->apm: %p %d %d\n", inst->apm, inst->sample_rate_hz, num_channels);

        float* channel_ptrs[1] = { samples };
        int32_t r = inst->apm->ProcessStream(channel_ptrs, stream_cfg, stream_cfg, channel_ptrs);

        //fprintf(stderr, "---ProcessStream: %d\n", r);
        return r;
    }

    int32_t apm_process_stream_int16(ApmHandle h, int16_t* samples, int32_t num_channels)
    {
        auto* inst = static_cast<ApmInstance*>(h);
        webrtc::StreamConfig stream_cfg(inst->sample_rate_hz, num_channels);

        return inst->apm->ProcessStream((const int16_t* const)samples, stream_cfg, stream_cfg, (int16_t* const)samples);
    }


    void apm_get_stats(ApmHandle h, ApmStats* out) {
        if (!h || !out) return;
        auto* inst = static_cast<ApmInstance*>(h);
        auto stats = inst->apm->GetStatistics();


        out->has_erl = stats.echo_return_loss.has_value() ? 1 : 0;
        out->echo_return_loss = stats.echo_return_loss.value_or(0.0);

        out->has_erle = stats.echo_return_loss_enhancement.has_value() ? 1 : 0;
        out->echo_return_loss_enhancement = stats.echo_return_loss_enhancement.value_or(0.0);

        out->has_divergent = stats.divergent_filter_fraction.has_value() ? 1 : 0;
        out->divergent_filter_fraction = stats.divergent_filter_fraction.value_or(0.0);

        out->has_delay = stats.delay_ms.has_value() ? 1 : 0;
        out->delay_ms = stats.delay_ms.value_or(0);

        out->has_residual_echo = stats.residual_echo_likelihood.has_value() ? 1 : 0;
        out->residual_echo_likelihood = stats.residual_echo_likelihood.value_or(0.0);
    }


#ifdef __cplusplus
}
#endif
