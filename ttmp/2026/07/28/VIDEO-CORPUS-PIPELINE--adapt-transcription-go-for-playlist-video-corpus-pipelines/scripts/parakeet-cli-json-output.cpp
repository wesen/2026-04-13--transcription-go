#include "parakeet.h"
#include "common-whisper.h"

#include <cstdio>
#include <string>
#include <thread>
#include <vector>
#include <cstring>
#include <fstream>
#include <iomanip>
#include <sstream>

// command-line parameters
struct parakeet_params {
    int32_t n_threads         = std::min(4, (int32_t) std::thread::hardware_concurrency());

    bool use_gpu       = true;
    int32_t gpu_device = 0;

    bool print_segments = false;
    bool output_txt     = false;
    bool output_json    = false;
    bool no_prints      = false;

    std::string model       = "models/ggml-parakeet-tdt-0.6b-v3.bin";
    std::string output_file = "";
    std::vector<std::string> fname_inp = {};
};

static void parakeet_print_usage(int argc, char ** argv, const parakeet_params & params);

static char * requires_value_error(const std::string & arg) {
    fprintf(stderr, "error: argument %s requires value\n", arg.c_str());
    exit(1);
}

static bool parakeet_params_parse(int argc, char ** argv, parakeet_params & params) {
    if (const char * env_device = std::getenv("PARAKEET_ARG_DEVICE")) {
        params.gpu_device = std::stoi(env_device);
    }

    for (int i = 1; i < argc; i++) {
        std::string arg = argv[i];

        if (arg == "-"){
            params.fname_inp.push_back(arg);
            continue;
        }

        if (arg[0] != '-') {
            params.fname_inp.push_back(arg);
            continue;
        }

        if (arg == "-h" || arg == "--help") {
            parakeet_print_usage(argc, argv, params);
            exit(0);
        }
        #define ARGV_NEXT (((i + 1) < argc) ? argv[++i] : requires_value_error(arg))
        else if (arg == "-t"    || arg == "--threads")         { params.n_threads         = std::stoi(ARGV_NEXT); }
        else if (arg == "-m"    || arg == "--model")           { params.model             = ARGV_NEXT; }
        else if (arg == "-f"    || arg == "--file")            { params.fname_inp.emplace_back(ARGV_NEXT); }
        else if (arg == "-ng"   || arg == "--no-gpu")          { params.use_gpu           = false; }
        else if (arg == "-dev"  || arg == "--device")          { params.gpu_device        = std::stoi(ARGV_NEXT); }
        else if (arg == "-ps"   || arg == "--print-segments")  { params.print_segments    = true; }
        else if (arg == "-otxt" || arg == "--output-txt")      { params.output_txt        = true; }
        else if (arg == "-oj"   || arg == "--output-json")     { params.output_json       = true; }
        else if (arg == "-of"   || arg == "--output-file")     { params.output_file       = ARGV_NEXT; }
        else if (arg == "-np"   || arg == "--no-prints")       { params.no_prints         = true; }
        else {
            fprintf(stderr, "error: unknown argument: %s\n", arg.c_str());
            parakeet_print_usage(argc, argv, params);
            exit(1);
        }
    }

    return true;
}

static void parakeet_print_usage(int /*argc*/, char ** argv, const parakeet_params & params) {
    fprintf(stderr, "\n");
    fprintf(stderr, "usage: %s [options] file0 file1 ...\n", argv[0]);
    fprintf(stderr, "supported audio formats: flac, mp3, ogg, wav\n");
    fprintf(stderr, "\n");
    fprintf(stderr, "options:\n");
    fprintf(stderr, "  -h,     --help              [default] show this help message and exit\n");
    fprintf(stderr, "  -t N,   --threads N         [%-7d] number of threads to use during computation\n", params.n_threads);
    fprintf(stderr, "  -m,     --model FILE        [%-7s] model path\n",                                  params.model.c_str());
    fprintf(stderr, "  -f N,   --file FILE         [%-7s] input audio file\n",                            "");
    fprintf(stderr, "  -ng,    --no-gpu            [%-7s] disable GPU\n",                                 params.use_gpu ? "false" : "true");
    fprintf(stderr, "  -dev N, --device N          [%-7d] GPU device to use\n",                           params.gpu_device);
    fprintf(stderr, "  -ps,    --print-segments    [%-7s] print segment information\n",                   params.print_segments ? "true" : "false");
    fprintf(stderr, "  -otxt,  --output-txt        [%-7s] output result in a text file\n",                params.output_txt ? "true" : "false");
    fprintf(stderr, "  -oj,    --output-json       [%-7s] output result to a JSON file with word timestamps\n", params.output_json ? "true" : "false");
    fprintf(stderr, "  -of,    --output-file FILE  [%-7s] output file path (without file extension)\n",   "");
    fprintf(stderr, "  -np,    --no-prints         [%-7s] do not print anything other than the results\n", params.no_prints ? "true" : "false");
    fprintf(stderr, "\n");
}

// Escape a string for JSON output.
static std::string json_escape(const std::string & s) {
    std::string out;
    out.reserve(s.size() + 8);
    for (size_t i = 0; i < s.size(); i++) {
        unsigned char c = (unsigned char)s[i];
        if (c == '"')       out += "\\\"";
        else if (c == '\\') out += "\\\\";
        else if (c == '\n') out += "\\n";
        else if (c == '\r') out += "\\r";
        else if (c == '\t') out += "\\t";
        else if (c < 0x20)  { std::ostringstream oss; oss << "\\u" << std::hex << std::setw(4) << std::setfill('0') << (int)c; out += oss.str(); }
        else                out += (char)c;
    }
    return out;
}

// Strip the SentencePiece "▁" (U+2581, UTF-8: E2 96 81) prefix from a token string.
static std::string strip_sp_marker(const std::string & s) {
    if (s.size() >= 3 && (unsigned char)s[0] == 0xE2 && (unsigned char)s[1] == 0x96 && (unsigned char)s[2] == 0x81) {
        return s.substr(3);
    }
    return s;
}

void token_callback(parakeet_context * ctx, parakeet_state * state, const parakeet_token_data * token_data, void * user_data) {
    bool * is_first = (bool *) user_data;

    const char * token_str = parakeet_token_to_str(ctx, token_data->id);
    char text_buf[256];
    parakeet_token_to_text(token_str, *is_first, text_buf, sizeof(text_buf));
    printf("%s", text_buf);
    fflush(stdout);

    *is_first = false;
}

static void cb_log_disable(enum ggml_log_level , const char * , void * ) { }

int main(int argc, char ** argv) {
    ggml_backend_load_all();

    parakeet_params params;

    if (parakeet_params_parse(argc, argv, params) == false) {
        return 1;
    }

    if (params.no_prints) {
        parakeet_log_set(cb_log_disable, NULL);
    }

    if (params.fname_inp.empty()) {
        fprintf(stderr, "error: no input files specified\n");
        parakeet_print_usage(argc, argv, params);
        return 1;
    }

    struct parakeet_context_params ctx_params = parakeet_context_default_params();
    ctx_params.use_gpu     = params.use_gpu;
    ctx_params.gpu_device  = params.gpu_device;

    if (!params.no_prints) {
        fprintf(stderr, "Loading Parakeet model from: %s\n", params.model.c_str());
    }


    struct parakeet_context * pctx = parakeet_init_from_file_with_params(params.model.c_str(), ctx_params);
    if (pctx == nullptr) {
        fprintf(stderr, "error: failed to load Parakeet model from '%s'\n", params.model.c_str());
        return 1;
    }

    if (!params.no_prints) {
        fprintf(stderr, "Successfully loaded Parakeet model\n");
        fprintf(stderr, "system_info: n_threads = %d / %d | %s\n",
                params.n_threads, (int32_t) std::thread::hardware_concurrency(), parakeet_print_system_info());
    }

    // Process each input file
    for (const auto & fname : params.fname_inp) {
        if (!params.no_prints) {
            fprintf(stderr, "\nProcessing file: %s\n", fname.c_str());
        }

        std::vector<float> pcmf32;
        std::vector<std::vector<float>> pcmf32s;
        if (!read_audio_data(fname.c_str(), pcmf32, pcmf32s, false)) {
            fprintf(stderr, "error: failed to read audio file '%s'\n", fname.c_str());
            continue;
        }

        if (pcmf32.empty()) {
            fprintf(stderr, "error: no audio data in file '%s'\n", fname.c_str());
            continue;
        }

        bool is_first = true;
        struct parakeet_full_params full_params = parakeet_full_default_params(PARAKEET_SAMPLING_GREEDY);
        full_params.n_threads           = params.n_threads;
        full_params.new_token_callback  = token_callback;
        full_params.new_token_callback_user_data = &is_first;

        const int mel_frames = (int)(pcmf32.size() / PARAKEET_HOP_LENGTH);
        int ret = parakeet_full(pctx, full_params, pcmf32.data(), pcmf32.size());

        if (ret != 0) {
            fprintf(stderr, "error: failed to process audio file '%s'\n", fname.c_str());
            continue;
        }

        printf("\n");

        if (params.output_txt) {
            const std::string fname_out = (!params.output_file.empty() ? params.output_file : fname) + ".txt";

            std::ofstream fout(fname_out);
            if (fout.is_open()) {
                const int n_segments = parakeet_full_n_segments(pctx);
                for (int i = 0; i < n_segments; ++i) {
                    const char * text = parakeet_full_get_segment_text(pctx, i);
                    fout << text << "\n";
                }
                fout.close();
                if (!params.no_prints) {
                    fprintf(stderr, "Output written to: %s\n", fname_out.c_str());
                }
            } else {
                fprintf(stderr, "error: failed to open '%s' for writing\n", fname_out.c_str());
            }
        }

        if (params.output_json) {
            const std::string fname_out = (!params.output_file.empty() ? params.output_file : fname) + ".json";

            std::ofstream fout(fname_out);
            if (fout.is_open()) {
                fout << std::fixed << std::setprecision(3);
                fout << "{\"transcription\":{\"segments\":[";

                const int n_segments = parakeet_full_n_segments(pctx);
                for (int i = 0; i < n_segments; ++i) {
                    const char * text = parakeet_full_get_segment_text(pctx, i);
                    const int64_t seg_t0 = parakeet_full_get_segment_t0(pctx, i);
                    const int64_t seg_t1 = parakeet_full_get_segment_t1(pctx, i);
                    const int n_tokens = parakeet_full_n_tokens(pctx, i);

                    if (i > 0) fout << ",";

                    // Parakeet timestamps are in mel-frame units; 80ms per frame.
                    fout << "{\"start\":" << (seg_t0 * 0.08)
                         << ",\"end\":" << (seg_t1 * 0.08)
                         << ",\"text\":\"" << json_escape(text ? text : "") << "\""
                         << ",\"words\":[";

                    // Accumulate subword tokens into complete words.
                    // Parakeet uses BPE: a word may span multiple tokens.
                    // is_word_start=true marks the first subword of each word.
                    // We accumulate text from word_start until the next word_start (or end),
                    // using the first token's t0 as the word start and the last token's t1 as end.
                    bool first_word = true;
                    std::string current_word;
                    double word_t0 = 0.0;
                    double word_t1 = 0.0;
                    bool in_word = false;

                    for (int j = 0; j < n_tokens; j++) {
                        parakeet_token_data td = parakeet_full_get_token_data(pctx, i, j);
                        const char * token_str = parakeet_token_to_str(pctx, td.id);
                        if (!token_str || !*token_str) continue;

                        if (td.is_word_start) {
                            // Flush previous word
                            if (in_word && !current_word.empty()) {
                                if (!first_word) fout << ",";
                                first_word = false;
                                fout << "{\"word\":\"" << json_escape(current_word) << "\""
                                     << ",\"start\":" << word_t0
                                     << ",\"end\":" << word_t1
                                     << "}";
                            }
                            // Start new word
                            current_word = strip_sp_marker(std::string(token_str));
                            word_t0 = td.t0 * 0.01;
                            word_t1 = td.t1 * 0.01;
                            in_word = true;
                        } else if (in_word) {
                            // Append subword continuation
                            current_word += strip_sp_marker(std::string(token_str));
                            word_t1 = td.t1 * 0.01;
                        }
                    }
                    // Flush last word
                    if (in_word && !current_word.empty()) {
                        if (!first_word) fout << ",";
                        fout << "{\"word\":\"" << json_escape(current_word) << "\""
                             << ",\"start\":" << word_t0
                             << ",\"end\":" << word_t1
                             << "}";
                    }

                    fout << "]}";
                }

                fout << "]}}\n";
                fout.close();

                if (!params.no_prints) {
                    fprintf(stderr, "Output written to: %s\n", fname_out.c_str());
                }
            } else {
                fprintf(stderr, "error: failed to open '%s' for writing\n", fname_out.c_str());
            }
        }

        if (!params.no_prints) {
            parakeet_print_timings(pctx);
        }

        if (params.print_segments) {
            const int n_segments = parakeet_full_n_segments(pctx);
            fprintf(stderr, "\nSegments (%d):\n", n_segments);

            for (int i = 0; i < n_segments; i++) {
                const char * text = parakeet_full_get_segment_text(pctx, i);
                const int64_t t0 = parakeet_full_get_segment_t0(pctx, i);
                const int64_t t1 = parakeet_full_get_segment_t1(pctx, i);
                const int n_tokens = parakeet_full_n_tokens(pctx, i);

                fprintf(stderr, "Segment %d: [%lld -> %lld] \"%s\"\n", i, (long long)t0, (long long)t1, text);
                fprintf(stderr, "Tokens [%d]:\n", n_tokens);

                for (int j = 0; j < n_tokens; j++) {
                    parakeet_token_data token_data = parakeet_full_get_token_data(pctx, i, j);
                    const char * token_str = parakeet_token_to_str(pctx, token_data.id);

                    fprintf(stderr, "  [%2d] id=%5d frame=%3d dur_idx=%2d dur_val=%2d p=%.4f plog=%.4f t0=%4lld t1=%4lld word_start=%s \"%s\"\n",
                           j,
                           token_data.id,
                           token_data.frame_index,
                           token_data.duration_idx,
                           token_data.duration_value,
                           token_data.p,
                           token_data.plog,
                           (long long)token_data.t0,
                           (long long)token_data.t1,
                           token_data.is_word_start ? "true": "false",
                           token_str);
                }
            }
        }
    }

    parakeet_free(pctx);

    return 0;
}
