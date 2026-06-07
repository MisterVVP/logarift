package llmqueue

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	DefaultStream   = "logarift:llm_enrichment_jobs"
	DefaultGroup    = "logarift-backend"
	DefaultConsumer = "backend-1"
)

type ValkeyStream struct {
	address  string
	password string
	database int
	stream   string
	group    string
	consumer string
	logger   *slog.Logger
}

type Options struct {
	URL      string
	Stream   string
	Group    string
	Consumer string
	Logger   *slog.Logger
}

func NewValkeyStream(opts Options) (*ValkeyStream, error) {
	if strings.TrimSpace(opts.URL) == "" {
		return nil, errors.New("valkey URL is required")
	}
	parsed, err := url.Parse(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("parse valkey URL: %w", err)
	}
	if parsed.Scheme != "valkey" && parsed.Scheme != "redis" {
		return nil, fmt.Errorf("valkey URL must use valkey:// or redis://, got %q", parsed.Scheme)
	}
	address := parsed.Host
	if !strings.Contains(address, ":") {
		address = net.JoinHostPort(address, "6379")
	}
	database := 0
	if path := strings.Trim(parsed.Path, "/"); path != "" {
		database, err = strconv.Atoi(path)
		if err != nil || database < 0 {
			return nil, fmt.Errorf("valkey database path must be a non-negative integer, got %q", path)
		}
	}
	stream := valueOrDefault(opts.Stream, DefaultStream)
	group := valueOrDefault(opts.Group, DefaultGroup)
	consumer := valueOrDefault(opts.Consumer, DefaultConsumer)
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	password, _ := parsed.User.Password()
	return &ValkeyStream{address: address, password: password, database: database, stream: stream, group: group, consumer: consumer, logger: logger}, nil
}

func (q *ValkeyStream) Enqueue(ctx context.Context, jobID bson.ObjectID) error {
	if jobID.IsZero() {
		return errors.New("job id is required")
	}
	if err := q.ensureGroup(ctx); err != nil {
		return err
	}
	_, err := q.command(ctx, "XADD", q.stream, "*", "job_id", jobID.Hex())
	if err != nil {
		return fmt.Errorf("xadd llm enrichment job: %w", err)
	}
	q.logger.Info("llm enrichment job published to valkey stream", "stream", q.stream, "group", q.group, "job_id", jobID.Hex())
	return nil
}

func (q *ValkeyStream) Start(ctx context.Context, handler func(context.Context, bson.ObjectID)) {
	if handler == nil {
		return
	}
	if err := q.ensureGroup(ctx); err != nil {
		q.logger.Error("failed to initialize valkey stream group", "stream", q.stream, "group", q.group, "error", err)
	}
	go q.consumeLoop(ctx, handler)
}

func (q *ValkeyStream) consumeLoop(ctx context.Context, handler func(context.Context, bson.ObjectID)) {
	for ctx.Err() == nil {
		jobID, messageID, err := q.read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			q.logger.Warn("failed to read valkey llm enrichment stream", "stream", q.stream, "group", q.group, "error", err)
			time.Sleep(time.Second)
			continue
		}
		if jobID.IsZero() {
			continue
		}
		handler(ctx, jobID)
		if messageID != "" {
			if err := q.ack(ctx, messageID); err != nil {
				q.logger.Warn("failed to ack valkey llm enrichment message", "stream", q.stream, "group", q.group, "message_id", messageID, "job_id", jobID.Hex(), "error", err)
			}
		}
	}
}

func (q *ValkeyStream) ensureGroup(ctx context.Context) error {
	_, err := q.command(ctx, "XGROUP", "CREATE", q.stream, q.group, "0", "MKSTREAM")
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create valkey stream group: %w", err)
	}
	return nil
}

func (q *ValkeyStream) read(ctx context.Context) (bson.ObjectID, string, error) {
	resp, err := q.command(ctx, "XREADGROUP", "GROUP", q.group, q.consumer, "BLOCK", "5000", "COUNT", "1", "STREAMS", q.stream, ">")
	if err != nil {
		return bson.NilObjectID, "", err
	}
	return parseJobMessage(resp)
}

func (q *ValkeyStream) ack(ctx context.Context, messageID string) error {
	_, err := q.command(ctx, "XACK", q.stream, q.group, messageID)
	return err
}

func (q *ValkeyStream) command(ctx context.Context, args ...string) (any, error) {
	deadline := time.Now().Add(7 * time.Second)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", q.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(deadline)
	reader := bufio.NewReader(conn)
	if q.password != "" {
		if _, err := writeCommand(conn, "AUTH", q.password); err != nil {
			return nil, err
		}
		if _, err := readRESP(reader); err != nil {
			return nil, err
		}
	}
	if q.database > 0 {
		if _, err := writeCommand(conn, "SELECT", strconv.Itoa(q.database)); err != nil {
			return nil, err
		}
		if _, err := readRESP(reader); err != nil {
			return nil, err
		}
	}
	if _, err := writeCommand(conn, args...); err != nil {
		return nil, err
	}
	return readRESP(reader)
}

func writeCommand(w io.Writer, args ...string) (int, error) {
	var b strings.Builder
	b.WriteString("*")
	b.WriteString(strconv.Itoa(len(args)))
	b.WriteString("\r\n")
	for _, arg := range args {
		b.WriteString("$")
		b.WriteString(strconv.Itoa(len(arg)))
		b.WriteString("\r\n")
		b.WriteString(arg)
		b.WriteString("\r\n")
	}
	return io.WriteString(w, b.String())
}

func readRESP(r *bufio.Reader) (any, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch prefix {
	case '+':
		line, err := readLine(r)
		return line, err
	case '-':
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(line)
	case ':':
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		return strconv.ParseInt(line, 10, 64)
	case '$':
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(line)
		if err != nil {
			return nil, err
		}
		if length < 0 {
			return nil, nil
		}
		buf := make([]byte, length+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return string(buf[:length]), nil
	case '*':
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(line)
		if err != nil {
			return nil, err
		}
		if length < 0 {
			return nil, nil
		}
		out := make([]any, 0, length)
		for i := 0; i < length; i++ {
			item, err := readRESP(r)
			if err != nil {
				return nil, err
			}
			out = append(out, item)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported RESP prefix %q", prefix)
	}
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}

func parseJobMessage(resp any) (bson.ObjectID, string, error) {
	streams, ok := resp.([]any)
	if !ok || len(streams) == 0 {
		return bson.NilObjectID, "", nil
	}
	streamEntry, ok := streams[0].([]any)
	if !ok || len(streamEntry) < 2 {
		return bson.NilObjectID, "", nil
	}
	messages, ok := streamEntry[1].([]any)
	if !ok || len(messages) == 0 {
		return bson.NilObjectID, "", nil
	}
	message, ok := messages[0].([]any)
	if !ok || len(message) < 2 {
		return bson.NilObjectID, "", nil
	}
	messageID, _ := message[0].(string)
	fields, ok := message[1].([]any)
	if !ok {
		return bson.NilObjectID, messageID, nil
	}
	for i := 0; i+1 < len(fields); i += 2 {
		name, _ := fields[i].(string)
		value, _ := fields[i+1].(string)
		if name == "job_id" {
			id, err := bson.ObjectIDFromHex(value)
			if err != nil {
				return bson.NilObjectID, messageID, err
			}
			return id, messageID, nil
		}
	}
	return bson.NilObjectID, messageID, nil
}

func valueOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
