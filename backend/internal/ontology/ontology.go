package ontology

var WorkflowStages = []string{"planning", "local_development", "build", "test", "code_review", "merge", "deployment", "operation", "debugging", "documentation", "coordination", "learning"}
var FrictionLayers = []string{"technical", "temporal", "cognitive", "social_process", "motivational", "environmental"}
var FrictionTypes = []string{"slow_feedback", "failed_feedback", "unclear_error", "missing_documentation", "ambiguous_ownership", "interruption", "waiting_for_review", "waiting_for_ci", "context_switch", "rework", "tooling_failure", "environment_setup", "coordination_overhead", "decision_blocked"}
var EventSources = []string{"manual", "seed", "import"}
var GoalStatuses = []string{"active", "completed", "deferred", "abandoned"}

const (
	SourceManual     = "manual"
	GoalStatusActive = "active"
)

var workflowStageSet = set(WorkflowStages...)
var frictionLayerSet = set(FrictionLayers...)
var frictionTypeSet = set(FrictionTypes...)
var eventSourceSet = set(EventSources...)
var goalStatusSet = set(GoalStatuses...)

func set(values ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}
func contains(m map[string]struct{}, value string) bool { _, ok := m[value]; return ok }
func IsWorkflowStage(value string) bool                 { return contains(workflowStageSet, value) }
func IsFrictionLayer(value string) bool                 { return contains(frictionLayerSet, value) }
func IsFrictionType(value string) bool                  { return contains(frictionTypeSet, value) }
func IsEventSource(value string) bool                   { return contains(eventSourceSet, value) }
func IsGoalStatus(value string) bool                    { return contains(goalStatusSet, value) }
