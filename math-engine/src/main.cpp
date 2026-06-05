#include <algorithm>
#include <atomic>
#include <chrono>
#include <arpa/inet.h>
#include <cerrno>
#include <cmath>
#include <csignal>
#include <cstdlib>
#include <cstring>
#include <ctime>
#include <cctype>
#include <iomanip>
#include <iostream>
#include <map>
#include <netinet/in.h>
#include <sstream>
#include <string>
#include <sys/socket.h>
#include <unistd.h>
#include <vector>

struct Event {
    std::string id;
    std::string timestamp_start;
    std::string timestamp_end;
    std::string friction_type;
    int severity_self = 0;
    int cognitive_load_self = 0;
    int time_lost_minutes = 0;
    int resume_time_minutes = 0;
    int recovery_minutes = 0;
    int interruption_count = 0;
};

static volatile std::sig_atomic_t stop_requested = 0;
static std::atomic<unsigned long long> request_counter{0};

static void handle_signal(int) {
    stop_requested = 1;
}

static std::string read_all_stdin() {
    std::ostringstream buffer;
    buffer << std::cin.rdbuf();
    return buffer.str();
}

static std::string escape_json(const std::string &value) {
    std::string out;
    out.reserve(value.size() + 8);
    for (char ch : value) {
        switch (ch) {
            case '\\': out += "\\\\"; break;
            case '"': out += "\\\""; break;
            case '\n': out += "\\n"; break;
            case '\r': out += "\\r"; break;
            case '\t': out += "\\t"; break;
            default: out += ch; break;
        }
    }
    return out;
}

static std::string error_json(const std::string &code, const std::string &message) {
    return "{\"error\":{\"code\":\"" + escape_json(code) + "\",\"message\":\"" + escape_json(message) + "\"}}\n";
}

static std::string utc_now_iso() {
    std::time_t now = std::time(nullptr);
    std::tm tm{};
#if defined(_WIN32)
    gmtime_s(&tm, &now);
#else
    gmtime_r(&now, &tm);
#endif
    char buffer[32];
    std::strftime(buffer, sizeof(buffer), "%Y-%m-%dT%H:%M:%SZ", &tm);
    return std::string(buffer);
}

static std::string format_double(double value) {
    std::ostringstream out;
    out << std::fixed << std::setprecision(4) << value;
    return out.str();
}

static void log_json(const std::string &level, const std::string &message,
                     const std::map<std::string, std::string> &fields = {}) {
    std::ostringstream out;
    out << "{\"timestamp\":\"" << utc_now_iso() << "\",\"level\":\"" << escape_json(level)
        << "\",\"service\":\"logarift-math-engine\",\"message\":\"" << escape_json(message) << "\"";
    for (const auto &entry : fields) {
        out << ",\"" << escape_json(entry.first) << "\":\"" << escape_json(entry.second) << "\"";
    }
    out << "}";
    std::cerr << out.str() << std::endl;
}

static size_t skip_ws(const std::string &s, size_t pos) {
    while (pos < s.size() && (s[pos] == ' ' || s[pos] == '\n' || s[pos] == '\r' || s[pos] == '\t')) {
        ++pos;
    }
    return pos;
}

static std::string find_string(const std::string &json, const std::string &key, const std::string &fallback = "") {
    const std::string needle = "\"" + key + "\"";
    size_t pos = json.find(needle);
    if (pos == std::string::npos) {
        return fallback;
    }
    pos = json.find(':', pos + needle.size());
    if (pos == std::string::npos) {
        return fallback;
    }
    pos = skip_ws(json, pos + 1);
    if (pos >= json.size() || json[pos] != '"') {
        return fallback;
    }
    ++pos;
    std::string out;
    bool escaped = false;
    for (; pos < json.size(); ++pos) {
        char ch = json[pos];
        if (escaped) {
            switch (ch) {
                case 'n': out += '\n'; break;
                case 'r': out += '\r'; break;
                case 't': out += '\t'; break;
                default: out += ch; break;
            }
            escaped = false;
            continue;
        }
        if (ch == '\\') {
            escaped = true;
            continue;
        }
        if (ch == '"') {
            return out;
        }
        out += ch;
    }
    return fallback;
}

static int find_int(const std::string &json, const std::string &key, int fallback = 0) {
    const std::string needle = "\"" + key + "\"";
    size_t pos = json.find(needle);
    if (pos == std::string::npos) {
        return fallback;
    }
    pos = json.find(':', pos + needle.size());
    if (pos == std::string::npos) {
        return fallback;
    }
    pos = skip_ws(json, pos + 1);
    bool neg = false;
    if (pos < json.size() && json[pos] == '-') {
        neg = true;
        ++pos;
    }
    int value = 0;
    bool any = false;
    while (pos < json.size() && json[pos] >= '0' && json[pos] <= '9') {
        any = true;
        value = value * 10 + (json[pos] - '0');
        ++pos;
    }
    if (!any) {
        return fallback;
    }
    return neg ? -value : value;
}

static size_t matching_brace(const std::string &json, size_t open) {
    int depth = 0;
    bool in_string = false;
    bool escaped = false;
    for (size_t i = open; i < json.size(); ++i) {
        char ch = json[i];
        if (in_string) {
            if (escaped) {
                escaped = false;
            } else if (ch == '\\') {
                escaped = true;
            } else if (ch == '"') {
                in_string = false;
            }
            continue;
        }
        if (ch == '"') {
            in_string = true;
        } else if (ch == '{') {
            ++depth;
        } else if (ch == '}') {
            --depth;
            if (depth == 0) {
                return i;
            }
        }
    }
    return std::string::npos;
}

static std::vector<Event> parse_events(const std::string &json) {
    std::vector<Event> events;
    size_t events_pos = json.find("\"events\"");
    if (events_pos == std::string::npos) {
        return events;
    }
    size_t array_open = json.find('[', events_pos);
    if (array_open == std::string::npos) {
        return events;
    }
    size_t pos = array_open + 1;
    while (pos < json.size()) {
        pos = json.find('{', pos);
        if (pos == std::string::npos) {
            break;
        }
        size_t end = matching_brace(json, pos);
        if (end == std::string::npos) {
            break;
        }
        std::string object = json.substr(pos, end - pos + 1);
        Event e;
        e.id = find_string(object, "id");
        e.timestamp_start = find_string(object, "timestamp_start");
        e.timestamp_end = find_string(object, "timestamp_end");
        e.friction_type = find_string(object, "friction_type");
        e.severity_self = find_int(object, "severity_self");
        e.cognitive_load_self = find_int(object, "cognitive_load_self");
        e.time_lost_minutes = find_int(object, "time_lost_minutes");
        e.resume_time_minutes = find_int(object, "resume_time_minutes");
        e.recovery_minutes = find_int(object, "recovery_minutes");
        e.interruption_count = find_int(object, "interruption_count");
        events.push_back(e);
        pos = end + 1;
        size_t array_close = json.find(']', pos);
        size_t next_open = json.find('{', pos);
        if (array_close != std::string::npos && (next_open == std::string::npos || array_close < next_open)) {
            break;
        }
    }
    return events;
}

static std::string normalize_time(std::string value) {
    if (value.empty()) {
        return value;
    }
    size_t z = value.find('Z');
    size_t plus = value.find('+', 10);
    size_t minus = value.find('-', 10);
    size_t tz = std::min(z == std::string::npos ? value.size() : z,
                         std::min(plus == std::string::npos ? value.size() : plus,
                                  minus == std::string::npos ? value.size() : minus));
    std::string base = value.substr(0, tz);
    size_t dot = base.find('.');
    if (dot != std::string::npos) {
        base = base.substr(0, dot);
    }
    return base + "Z";
}

static std::time_t parse_time_utc(const std::string &value) {
    std::string normalized = normalize_time(value);
    if (normalized.empty()) {
        return 0;
    }
    std::tm tm{};
    std::istringstream ss(normalized);
    ss >> std::get_time(&tm, "%Y-%m-%dT%H:%M:%SZ");
    if (ss.fail()) {
        return 0;
    }
#if defined(_WIN32)
    return _mkgmtime(&tm);
#else
    return timegm(&tm);
#endif
}

static bool is_wait_like(const std::string &friction_type) {
    return friction_type == "waiting_for_review" || friction_type == "waiting_for_ci" ||
           friction_type == "decision_blocked" || friction_type == "coordination_overhead";
}

static double event_duration_minutes(const Event &e) {
    std::time_t start = parse_time_utc(e.timestamp_start);
    std::time_t end = parse_time_utc(e.timestamp_end);
    if (start > 0 && end >= start) {
        return std::max(0.0, std::difftime(end, start) / 60.0);
    }
    return static_cast<double>(std::max(0, e.time_lost_minutes));
}

struct EventScore {
    std::string id;
    double fcs = 0;
    std::string reason;
};

static std::string score_json(const std::string &input) {
    auto calculation_started = std::chrono::steady_clock::now();
    std::string period_start = find_string(input, "period_start");
    std::string period_end = find_string(input, "period_end");
    if (period_start.empty() || period_end.empty()) {
        log_json("warn", "score calculation rejected", {{"reason", "missing_period"}, {"payload_bytes", std::to_string(input.size())}});
        return error_json("invalid_input", "period_start and period_end are required");
    }

    std::vector<Event> events = parse_events(input);
    std::sort(events.begin(), events.end(), [](const Event &a, const Event &b) {
        return parse_time_utc(a.timestamp_start) < parse_time_utc(b.timestamp_start);
    });

    constexpr double decay = 0.85;
    constexpr double severity_multiplier = 1.2;
    constexpr double cognitive_multiplier = 1.5;
    constexpr double interruption_multiplier = 2.0;
    constexpr double recovery_multiplier = 0.3;
    constexpr double half_life_minutes = 90.0;

    double cla = 0.0;
    double fci = 0.0;
    double total_wait_time = 0.0;
    double total_active_time = 0.0;
    std::time_t period_end_time = parse_time_utc(period_end);
    std::vector<EventScore> event_scores;

    for (const Event &e : events) {
        double resume_penalty = std::log(1.0 + static_cast<double>(std::max(0, e.resume_time_minutes)));
        double recovery_bonus = static_cast<double>(std::max(0, e.recovery_minutes)) * recovery_multiplier;
        cla = std::max(0.0,
                       decay * cla + static_cast<double>(e.severity_self) * severity_multiplier +
                       static_cast<double>(e.cognitive_load_self) * cognitive_multiplier +
                       static_cast<double>(e.interruption_count) * interruption_multiplier + resume_penalty - recovery_bonus);

        std::time_t event_time = parse_time_utc(e.timestamp_start);
        double delta_minutes = 0.0;
        if (period_end_time > 0 && event_time > 0 && period_end_time >= event_time) {
            delta_minutes = std::difftime(period_end_time, event_time) / 60.0;
        }
        double event_weight = static_cast<double>(e.severity_self + e.cognitive_load_self) + resume_penalty;
        fci += event_weight * std::exp(-delta_minutes / half_life_minutes);

        double duration = event_duration_minutes(e);
        if (duration <= 0.0) {
            duration = static_cast<double>(std::max(0, e.time_lost_minutes));
        }
        total_active_time += duration;
        if (is_wait_like(e.friction_type)) {
            total_wait_time += static_cast<double>(std::max(0, e.time_lost_minutes));
        }

        double fcs = static_cast<double>(e.severity_self) *
                     (1.0 + std::log(1.0 + static_cast<double>(std::max(0, e.time_lost_minutes)))) *
                     (1.0 + 0.2 * static_cast<double>(std::max(0, e.interruption_count)));
        std::string reason = "high severity";
        if (e.resume_time_minutes >= 10) {
            reason = "high severity and long resume time";
        } else if (e.interruption_count > 0) {
            reason = "high severity with interruptions";
        } else if (is_wait_like(e.friction_type)) {
            reason = "wait-like friction type";
        }
        event_scores.push_back(EventScore{e.id, fcs, reason});
    }

    std::sort(event_scores.begin(), event_scores.end(), [](const EventScore &a, const EventScore &b) {
        return a.fcs > b.fcs;
    });

    double sdc = total_wait_time / std::max(total_active_time, 1.0);
    auto calculation_finished = std::chrono::steady_clock::now();
    auto duration_ms = std::chrono::duration_cast<std::chrono::milliseconds>(calculation_finished - calculation_started).count();
    log_json("info", "score calculation completed", {
        {"period_start", period_start},
        {"period_end", period_end},
        {"event_count", std::to_string(events.size())},
        {"top_contributor_count", std::to_string(std::min<size_t>(5, event_scores.size()))},
        {"total_wait_minutes", format_double(total_wait_time)},
        {"total_active_minutes", format_double(total_active_time)},
        {"cla", format_double(cla)},
        {"fci", format_double(fci)},
        {"sdc", format_double(sdc)},
        {"duration_ms", std::to_string(duration_ms)}
    });

    std::ostringstream out;
    out << std::fixed << std::setprecision(4);
    out << "{\n";
    out << "  \"model_version\": \"mvp-0.1\",\n";
    out << "  \"period_start\": \"" << escape_json(period_start) << "\",\n";
    out << "  \"period_end\": \"" << escape_json(period_end) << "\",\n";
    out << "  \"scores\": {\"cla\": " << cla << ", \"fci\": " << fci << ", \"sdc\": " << sdc << "},\n";
    out << "  \"event_scores\": [";
    for (size_t i = 0; i < event_scores.size(); ++i) {
        if (i) out << ", ";
        out << "{\"event_id\": \"" << escape_json(event_scores[i].id) << "\", \"fcs\": " << event_scores[i].fcs << "}";
    }
    out << "],\n";
    out << "  \"top_contributors\": [";
    size_t top_count = std::min<size_t>(5, event_scores.size());
    for (size_t i = 0; i < top_count; ++i) {
        if (i) out << ", ";
        out << "{\"event_id\": \"" << escape_json(event_scores[i].id) << "\", \"reason\": \"" << escape_json(event_scores[i].reason) << "\"}";
    }
    out << "]\n";
    out << "}\n";
    return out.str();
}

static std::string reason_phrase(int status) {
    switch (status) {
        case 200: return "OK";
        case 400: return "Bad Request";
        case 404: return "Not Found";
        case 405: return "Method Not Allowed";
        case 413: return "Payload Too Large";
        case 500: return "Internal Server Error";
        default: return "OK";
    }
}

static void send_http_response(int fd, int status, const std::string &body) {
    std::ostringstream response;
    response << "HTTP/1.1 " << status << " " << reason_phrase(status) << "\r\n";
    response << "Content-Type: application/json; charset=utf-8\r\n";
    response << "Content-Length: " << body.size() << "\r\n";
    response << "Connection: close\r\n";
    response << "\r\n";
    response << body;
    std::string data = response.str();
    const char *buf = data.data();
    size_t left = data.size();
    while (left > 0) {
        ssize_t sent = send(fd, buf, left, 0);
        if (sent <= 0) {
            if (errno == EINTR) {
                continue;
            }
            break;
        }
        buf += sent;
        left -= static_cast<size_t>(sent);
    }
}

static std::map<std::string, std::string> parse_headers(const std::string &header_block) {
    std::map<std::string, std::string> headers;
    std::istringstream input(header_block);
    std::string line;
    std::getline(input, line);
    while (std::getline(input, line)) {
        if (!line.empty() && line.back() == '\r') {
            line.pop_back();
        }
        if (line.empty()) {
            break;
        }
        size_t colon = line.find(':');
        if (colon == std::string::npos) {
            continue;
        }
        std::string key = line.substr(0, colon);
        std::string value = line.substr(colon + 1);
        while (!value.empty() && value.front() == ' ') {
            value.erase(value.begin());
        }
        std::transform(key.begin(), key.end(), key.begin(), [](unsigned char ch) { return static_cast<char>(std::tolower(ch)); });
        headers[key] = value;
    }
    return headers;
}

static void handle_client(int fd) {
    constexpr size_t max_request_size = 1024 * 1024;
    std::string request;
    char buffer[4096];
    size_t header_end = std::string::npos;
    while (request.size() < max_request_size) {
        ssize_t n = recv(fd, buffer, sizeof(buffer), 0);
        if (n < 0) {
            if (errno == EINTR) {
                continue;
            }
            log_json("warn", "HTTP request read failed", {{"error", std::strerror(errno)}});
            send_http_response(fd, 400, error_json("bad_request", "failed to read request"));
            return;
        }
        if (n == 0) {
            break;
        }
        request.append(buffer, static_cast<size_t>(n));
        header_end = request.find("\r\n\r\n");
        if (header_end != std::string::npos) {
            break;
        }
    }
    if (header_end == std::string::npos) {
        log_json("warn", "HTTP request rejected", {{"reason", "missing_headers"}});
        send_http_response(fd, 400, error_json("bad_request", "missing HTTP headers"));
        return;
    }

    std::string headers_raw = request.substr(0, header_end + 4);
    std::istringstream first_line(headers_raw);
    std::string method;
    std::string path;
    std::string version;
    first_line >> method >> path >> version;
    std::map<std::string, std::string> headers = parse_headers(headers_raw);
    size_t content_length = 0;
    auto cl = headers.find("content-length");
    if (cl != headers.end()) {
        try {
            content_length = static_cast<size_t>(std::stoul(cl->second));
        } catch (...) {
            send_http_response(fd, 400, error_json("bad_request", "invalid content-length"));
            return;
        }
    }
    if (content_length > max_request_size) {
        send_http_response(fd, 413, error_json("payload_too_large", "request body is too large"));
        return;
    }

    size_t body_start = header_end + 4;
    while (request.size() < body_start + content_length) {
        ssize_t n = recv(fd, buffer, sizeof(buffer), 0);
        if (n < 0) {
            if (errno == EINTR) {
                continue;
            }
            send_http_response(fd, 400, error_json("bad_request", "failed to read request body"));
            return;
        }
        if (n == 0) {
            break;
        }
        request.append(buffer, static_cast<size_t>(n));
    }
    std::string body;
    if (request.size() >= body_start) {
        body = request.substr(body_start, std::min(content_length, request.size() - body_start));
    }

    if ((path == "/health/live" || path == "/health/ready") && method == "GET") {
        send_http_response(fd, 200, "{\"status\":\"ok\",\"service\":\"logarift-math-engine\"}\n");
        return;
    }
    if (path == "/v1/score" && method == "POST") {
        unsigned long long request_id = ++request_counter;
        auto request_started = std::chrono::steady_clock::now();
        log_json("info", "score request received", {{"request_id", std::to_string(request_id)}, {"payload_bytes", std::to_string(body.size())}});
        std::string output = score_json(body);
        int status = output.find("\"error\"") != std::string::npos ? 400 : 200;
        auto request_finished = std::chrono::steady_clock::now();
        auto duration_ms = std::chrono::duration_cast<std::chrono::milliseconds>(request_finished - request_started).count();
        log_json(status == 200 ? "info" : "warn", "score request completed", {{"request_id", std::to_string(request_id)}, {"status", std::to_string(status)}, {"duration_ms", std::to_string(duration_ms)}});
        send_http_response(fd, status, output);
        return;
    }
    if (path == "/v1/score") {
        send_http_response(fd, 405, error_json("method_not_allowed", "use POST /v1/score"));
        return;
    }
    send_http_response(fd, 404, error_json("not_found", "endpoint not found"));
}

static int env_int(const char *name, int fallback) {
    const char *raw = std::getenv(name);
    if (raw == nullptr || *raw == '\0') {
        return fallback;
    }
    try {
        return std::stoi(raw);
    } catch (...) {
        return fallback;
    }
}


static int healthcheck() {
    int port = env_int("LOGARIFT_MATH_ENGINE_PORT", 8090);
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) {
        return 1;
    }
    sockaddr_in address{};
    address.sin_family = AF_INET;
    address.sin_port = htons(static_cast<uint16_t>(port));
    if (inet_pton(AF_INET, "127.0.0.1", &address.sin_addr) != 1) {
        close(fd);
        return 1;
    }
    if (connect(fd, reinterpret_cast<sockaddr *>(&address), sizeof(address)) < 0) {
        close(fd);
        return 1;
    }
    std::string request = "GET /health/ready HTTP/1.1\r\nHost: 127.0.0.1\r\nConnection: close\r\n\r\n";
    if (send(fd, request.data(), request.size(), 0) < 0) {
        close(fd);
        return 1;
    }
    char buffer[128];
    ssize_t n = recv(fd, buffer, sizeof(buffer) - 1, 0);
    close(fd);
    if (n <= 0) {
        return 1;
    }
    buffer[n] = '\0';
    std::string response(buffer);
    return response.find("HTTP/1.1 200") == 0 ? 0 : 1;
}

static int serve() {
    std::signal(SIGTERM, handle_signal);
    std::signal(SIGINT, handle_signal);

    int port = env_int("LOGARIFT_MATH_ENGINE_PORT", 8090);
    int server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (server_fd < 0) {
        std::cerr << "failed to create socket: " << std::strerror(errno) << "\n";
        return 1;
    }

    int reuse = 1;
    setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));

    sockaddr_in address{};
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = INADDR_ANY;
    address.sin_port = htons(static_cast<uint16_t>(port));

    if (bind(server_fd, reinterpret_cast<sockaddr *>(&address), sizeof(address)) < 0) {
        std::cerr << "failed to bind port " << port << ": " << std::strerror(errno) << "\n";
        close(server_fd);
        return 1;
    }
    if (listen(server_fd, 64) < 0) {
        std::cerr << "failed to listen: " << std::strerror(errno) << "\n";
        close(server_fd);
        return 1;
    }

    log_json("info", "math engine listening", {{"port", std::to_string(port)}});
    while (!stop_requested) {
        sockaddr_in client_address{};
        socklen_t client_len = sizeof(client_address);
        int client_fd = accept(server_fd, reinterpret_cast<sockaddr *>(&client_address), &client_len);
        if (client_fd < 0) {
            if (errno == EINTR) {
                continue;
            }
            log_json("error", "accept failed", {{"error", std::strerror(errno)}});
            continue;
        }
        handle_client(client_fd);
        close(client_fd);
    }
    log_json("info", "math engine stopped");
    close(server_fd);
    return 0;
}

int main(int argc, char **argv) {
    if (argc > 1 && std::string(argv[1]) == "--serve") {
        return serve();
    }
    if (argc > 1 && std::string(argv[1]) == "--healthcheck") {
        return healthcheck();
    }
    std::string output = score_json(read_all_stdin());
    std::cout << output;
    return output.find("\"error\"") != std::string::npos ? 1 : 0;
}
