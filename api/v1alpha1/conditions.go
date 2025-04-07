package v1alpha1

type ConditionType string

const (
	ConditionTypePending    ConditionType = ConditionType("Pending")
	ConditionTypeFailure    ConditionType = ConditionType("Failure")
	ConditionTypeInProgress ConditionType = ConditionType("InProgress")
	ConditionTypeInSync     ConditionType = ConditionType("InSync")
	ConditionTypeReady      ConditionType = ConditionType("Ready")
)

type ConditionReason string

const (
	ConditionReasonPending    ConditionReason = ConditionReason("Pending")
	ConditionReasonFailed     ConditionReason = ConditionReason("Failed")
	ConditionReasonInProgress ConditionReason = ConditionReason("InProgress")
	ConditionReasonInSync     ConditionReason = ConditionReason("InSync")
	ConditionReasonDone       ConditionReason = ConditionReason("Done")
)
