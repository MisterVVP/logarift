package enrichment

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/ontology"
)

const (
	EngineVersion = "rules-0.1"
	EngineType    = "rules"

	LevelGreen  = "green"
	LevelYellow = "yellow"
	LevelOrange = "orange"
	LevelRed    = "red"
)

type Input struct {
	OccurredAt    time.Time
	FrictionLevel string
	NotesMarkdown string
	Links         []string
	Attachments   []domain.FrictionAttachment
}

type Engine struct{}

func NewEngine() Engine { return Engine{} }

func ValidLevel(level string) bool {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case LevelGreen, LevelYellow, LevelOrange, LevelRed:
		return true
	default:
		return false
	}
}

func (Engine) Enrich(input Input, now time.Time) domain.FrictionEvent {
	level := strings.ToLower(strings.TrimSpace(input.FrictionLevel))
	plain := plainText(input.NotesMarkdown)
	lower := strings.ToLower(plain)
	links := mergeLinks(input.Links, extractLinks(input.NotesMarkdown))

	stage, stageConfidence, stageWhy := classifyWorkflowStage(lower, links)
	layer, layerConfidence, layerWhy := classifyFrictionLayer(lower)
	frictionType, typeConfidence, typeWhy := classifyFrictionType(lower, links)
	severity, cognitive, valence := scoresForLevel(level)
	timeLost, timeLostConfidence, timeLostWhy := inferTimeLost(level, lower)
	resumeTime, resumeConfidence, resumeWhy := inferResumeTime(level, lower)
	interruptions, interruptionConfidence, interruptionWhy := inferInterruptions(lower)
	tags := extractTags(lower, links)

	canonical := domain.FrictionCanonical{
		WorkflowStage:     stage,
		FrictionLayer:     layer,
		FrictionType:      frictionType,
		SeveritySelf:      severity,
		CognitiveLoadSelf: cognitive,
		EmotionValence:    valence,
		TimeLostMinutes:   timeLost,
		ResumeTimeMinutes: resumeTime,
		RecoveryMinutes:   0,
		InterruptionCount: interruptions,
		Tags:              tags,
	}

	fields := map[string]domain.FrictionFieldInference{
		"workflow_stage":      field(stage, stageConfidence, stageWhy),
		"friction_layer":      field(layer, layerConfidence, layerWhy),
		"friction_type":       field(frictionType, typeConfidence, typeWhy),
		"severity_self":       field(severity, 1.0, "Mapped directly from friction level."),
		"cognitive_load_self": field(cognitive, 0.82, "Estimated from friction level."),
		"emotion_valence":     field(valence, 0.82, "Estimated from friction level."),
		"time_lost_minutes":   field(timeLost, timeLostConfidence, timeLostWhy),
		"resume_time_minutes": field(resumeTime, resumeConfidence, resumeWhy),
		"interruption_count":  field(interruptions, interruptionConfidence, interruptionWhy),
		"tags":                field(tags, 0.80, "Extracted from known Developer Experience keywords and links."),
	}

	event := domain.FrictionEvent{
		SchemaVersion:     domain.CurrentSchemaVersion,
		InputMode:         "quick",
		TimestampStart:    input.OccurredAt.UTC(),
		WorkflowStage:     canonical.WorkflowStage,
		FrictionLayer:     canonical.FrictionLayer,
		FrictionType:      canonical.FrictionType,
		SeveritySelf:      canonical.SeveritySelf,
		CognitiveLoadSelf: canonical.CognitiveLoadSelf,
		EmotionValence:    canonical.EmotionValence,
		TimeLostMinutes:   canonical.TimeLostMinutes,
		ResumeTimeMinutes: canonical.ResumeTimeMinutes,
		RecoveryMinutes:   canonical.RecoveryMinutes,
		InterruptionCount: canonical.InterruptionCount,
		Tags:              canonical.Tags,
		Notes:             input.NotesMarkdown,
		Source:            ontology.SourceManual,
		CreatedAt:         now.UTC(),
		UpdatedAt:         now.UTC(),
		Observed: &domain.FrictionObserved{
			OccurredAt:    input.OccurredAt.UTC(),
			FrictionLevel: level,
			NotesMarkdown: input.NotesMarkdown,
			PlainText:     plain,
			Links:         links,
			Attachments:   input.Attachments,
		},
		Inference: &domain.FrictionInference{
			EngineVersion: EngineVersion,
			EngineType:    EngineType,
			CreatedAt:     now.UTC(),
			Fields:        fields,
		},
		Canonical: &canonical,
	}
	return event
}

func field(value any, confidence float64, explanation string) domain.FrictionFieldInference {
	return domain.FrictionFieldInference{Value: value, Confidence: clamp(confidence), Source: EngineType, Explanation: explanation}
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func plainText(markdown string) string {
	text := strings.ReplaceAll(markdown, "\r\n", "\n")
	text = regexp.MustCompile("[`*_>#]+").ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

var urlRe = regexp.MustCompile(`https?://[^\s)\]}>'"]+`)

func extractLinks(text string) []domain.FrictionLink {
	matches := urlRe.FindAllString(text, -1)
	out := make([]domain.FrictionLink, 0, len(matches))
	seen := map[string]struct{}{}
	for _, raw := range matches {
		url := strings.TrimRight(raw, ".,;:!")
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		out = append(out, domain.FrictionLink{URL: url, Source: "notes"})
	}
	return out
}

func mergeLinks(explicit []string, extracted []domain.FrictionLink) []domain.FrictionLink {
	seen := map[string]struct{}{}
	out := []domain.FrictionLink{}
	for _, raw := range explicit {
		url := strings.TrimSpace(raw)
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}
		out = append(out, domain.FrictionLink{URL: url, Source: "explicit"})
	}
	for _, link := range extracted {
		if _, ok := seen[link.URL]; ok {
			continue
		}
		seen[link.URL] = struct{}{}
		out = append(out, link)
	}
	return out
}

type classifierRule struct {
	value    string
	keywords []string
	why      string
}

func choose(defaultValue string, text string, rules []classifierRule) (string, float64, string) {
	bestValue := defaultValue
	bestScore := 0
	bestWhy := "No strong keyword match; defaulted to " + defaultValue + "."
	for _, rule := range rules {
		score := keywordScore(text, rule.keywords)
		if score > bestScore {
			bestScore = score
			bestValue = rule.value
			bestWhy = rule.why
		}
	}
	if bestScore == 0 {
		return bestValue, 0.42, bestWhy
	}
	confidence := 0.52 + float64(bestScore)*0.11
	if confidence > 0.92 {
		confidence = 0.92
	}
	return bestValue, confidence, bestWhy
}

func keywordScore(text string, keywords []string) int {
	score := 0
	for _, keyword := range keywords {
		if containsKeyword(text, keyword) {
			score++
		}
	}
	return score
}

func containsKeyword(text, keyword string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return false
	}
	if len(keyword) <= 2 && isAlphaNumericKeyword(keyword) {
		pattern := `(^|[^a-z0-9])` + regexp.QuoteMeta(keyword) + `([^a-z0-9]|$)`
		return regexp.MustCompile(pattern).MatchString(text)
	}
	return strings.Contains(text, keyword)
}

func isAlphaNumericKeyword(value string) bool {
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func linkText(links []domain.FrictionLink) string {
	var b strings.Builder
	for _, link := range links {
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(link.URL))
	}
	return b.String()
}

func classifyWorkflowStage(text string, links []domain.FrictionLink) (string, float64, string) {
	combined := text + linkText(links)
	return choose("local_development", combined, []classifierRule{
		{"test", []string{"ci", "pipeline", "test", "flaky", "github action", "actions", "pytest", "jest", "unit test", "integration test"}, "Matched testing, CI, pipeline, or flaky-test language."},
		{"build", []string{"build", "compile", "compiler", "cmake", "gradle", "maven", "npm install", "go build"}, "Matched build or compilation language."},
		{"code_review", []string{"pr", "pull request", "merge request", "review", "approval", "approver"}, "Matched code-review or approval language."},
		{"deployment", []string{"deploy", "deployment", "release", "helm", "k8s", "kubernetes", "production"}, "Matched deployment or release language."},
		{"debugging", []string{"debug", "debugging", "logs", "trace", "stacktrace", "investigate", "reproduce"}, "Matched debugging or investigation language."},
		{"documentation", []string{"docs", "documentation", "readme", "confluence", "wiki", "runbook"}, "Matched documentation language."},
		{"coordination", []string{"meeting", "slack", "handoff", "sync", "coordination", "ownership"}, "Matched coordination or communication language."},
		{"planning", []string{"requirement", "requirements", "spec", "scope", "planning", "story"}, "Matched planning, requirement, or scope language."},
		{"operation", []string{"incident", "alert", "oncall", "monitoring", "prod issue"}, "Matched operations or incident language."},
		{"learning", []string{"learn", "tutorial", "how to", "unknown library", "new library"}, "Matched learning or unfamiliar-tooling language."},
	})
}

func classifyFrictionLayer(text string) (string, float64, string) {
	return choose("technical", text, []classifierRule{
		{"technical", []string{"error", "failed", "failure", "crash", "broken", "timeout", "flaky", "bug", "exception"}, "Matched technical failure language."},
		{"temporal", []string{"slow", "wait", "waiting", "waited", "queue", "blocked", "delay", "delayed"}, "Matched waiting, delay, or slow-feedback language."},
		{"cognitive", []string{"unclear", "confusing", "ambiguous", "understand", "no idea", "complicated", "hard to"}, "Matched confusion or cognitive-load language."},
		{"social_process", []string{"ownership", "owner", "review", "approval", "decision", "handoff", "who owns"}, "Matched ownership, review, or decision-process language."},
		{"motivational", []string{"frustrating", "frustrated", "demotivating", "exhausting", "pointless", "annoying"}, "Matched frustration or motivation language."},
		{"environmental", []string{"network", "vpn", "laptop", "noise", "office", "interrupted", "interruption"}, "Matched environment or interruption language."},
	})
}

func classifyFrictionType(text string, links []domain.FrictionLink) (string, float64, string) {
	combined := text + linkText(links)
	if keywordScore(combined, []string{"ci", "pipeline", "github action", "actions"}) > 0 && keywordScore(combined, []string{"wait", "waiting", "queue", "queued"}) > 0 {
		return "waiting_for_ci", 0.88, "Matched CI/pipeline waiting language."
	}
	if keywordScore(combined, []string{"ci", "pipeline", "github action", "actions", "test"}) > 0 && keywordScore(combined, []string{"failed", "failure", "flaky", "red", "broken", "timeout"}) > 0 {
		return "failed_feedback", 0.90, "Matched CI/test failure language."
	}
	if keywordScore(combined, []string{"pr", "pull request", "merge request", "review"}) > 0 && keywordScore(combined, []string{"wait", "waiting", "blocked", "approval", "approver"}) > 0 {
		return "waiting_for_review", 0.90, "Matched review waiting language."
	}
	return choose("unclear_error", combined, []classifierRule{
		{"slow_feedback", []string{"slow", "took too long", "long feedback", "delayed feedback"}, "Matched slow-feedback language."},
		{"unclear_error", []string{"unclear", "confusing", "cryptic", "no idea", "unknown error", "hard to understand"}, "Matched unclear-error or confusing-feedback language."},
		{"missing_documentation", []string{"missing docs", "no docs", "documentation", "readme", "runbook", "confluence", "wiki"}, "Matched missing-documentation language."},
		{"ambiguous_ownership", []string{"ownership", "no owner", "nobody owns", "who owns", "unclear owner"}, "Matched ambiguous-ownership language."},
		{"interruption", []string{"interrupted", "interruption", "pinged", "meeting", "call"}, "Matched interruption language."},
		{"context_switch", []string{"context switch", "switched context", "lost context", "attention"}, "Matched context-switch language."},
		{"rework", []string{"rework", "redo", "redid", "rewrite", "start over"}, "Matched rework language."},
		{"tooling_failure", []string{"tooling", "tool failed", "cli failed", "ide", "plugin", "extension"}, "Matched tooling-failure language."},
		{"environment_setup", []string{"setup", "local env", "environment", "docker compose", "localhost", "devcontainer"}, "Matched environment-setup language."},
		{"coordination_overhead", []string{"handoff", "coordination", "sync", "meeting", "slack thread"}, "Matched coordination-overhead language."},
		{"decision_blocked", []string{"decision", "blocked on decision", "waiting for decision", "approval"}, "Matched decision-blocked language."},
	})
}

func scoresForLevel(level string) (int, int, int) {
	switch level {
	case LevelGreen:
		return 1, 1, 0
	case LevelYellow:
		return 2, 2, -1
	case LevelRed:
		return 5, 5, -2
	default:
		return 4, 4, -1
	}
}

var durationRe = regexp.MustCompile(`(?i)(\d+)\s*(hours?|hrs?|hr|h|minutes?|mins?|min|m)\b`)

func inferTimeLost(level, text string) (int, float64, string) {
	matches := durationRe.FindAllStringSubmatch(text, -1)
	best := 0
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		value := atoi(match[1])
		unit := strings.ToLower(match[2])
		minutes := value
		if strings.HasPrefix(unit, "h") {
			minutes = value * 60
		}
		if minutes > best {
			best = minutes
		}
	}
	if best > 0 {
		return best, 0.86, "Parsed explicit duration from notes."
	}
	return defaultTimeLost(level), 0.45, "No explicit duration found; estimated from friction level."
}

func atoi(value string) int {
	out := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return out
		}
		out = out*10 + int(r-'0')
	}
	return out
}

func defaultTimeLost(level string) int {
	switch level {
	case LevelGreen:
		return 2
	case LevelYellow:
		return 10
	case LevelRed:
		return 60
	default:
		return 30
	}
}

func inferResumeTime(level, text string) (int, float64, string) {
	resume := map[string]int{LevelGreen: 0, LevelYellow: 2, LevelOrange: 8, LevelRed: 15}[level]
	confidence := 0.42
	reason := "Estimated from friction level."
	if keywordScore(text, []string{"interrupted", "interruption", "context switch", "lost context", "meeting", "call", "pinged", "slack"}) > 0 {
		if resume < 10 {
			resume = 10
		}
		confidence = 0.70
		reason = "Estimated from interruption or context-switch language."
	}
	return resume, confidence, reason
}

func inferInterruptions(text string) (int, float64, string) {
	score := keywordScore(text, []string{"interrupted", "interruption", "context switch", "lost context", "meeting", "call", "pinged", "slack"})
	if score == 0 {
		return 0, 0.62, "No interruption language found."
	}
	if score > 3 {
		score = 3
	}
	return score, 0.76, "Count estimated from interruption-related keywords."
}

func extractTags(text string, links []domain.FrictionLink) []string {
	tags := map[string]struct{}{}
	addIf := func(tag string, keywords ...string) {
		if keywordScore(text, keywords) > 0 {
			tags[tag] = struct{}{}
		}
	}
	addIf("ci", "ci", "pipeline", "github action", "actions")
	addIf("flaky-test", "flaky")
	addIf("timeout", "timeout")
	addIf("review", "pr", "pull request", "review")
	addIf("docs", "docs", "documentation", "readme", "wiki", "confluence")
	addIf("docker", "docker", "docker compose", "container")
	addIf("kubernetes", "k8s", "kubernetes", "helm")
	addIf("debugging", "debug", "logs", "trace", "reproduce")
	addIf("interruption", "interrupted", "meeting", "slack", "pinged")
	for _, link := range links {
		url := strings.ToLower(link.URL)
		switch {
		case strings.Contains(url, "github"):
			tags["github"] = struct{}{}
		case strings.Contains(url, "jira"):
			tags["jira"] = struct{}{}
		case strings.Contains(url, "confluence") || strings.Contains(url, "wiki"):
			tags["docs"] = struct{}{}
		case strings.Contains(url, "slack"):
			tags["slack"] = struct{}{}
		}
	}
	out := make([]string, 0, len(tags))
	for tag := range tags {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}
