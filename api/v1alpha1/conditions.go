package v1alpha1

type ConditionType string

const (
	ConditionTypePending    ConditionType = ConditionType("Pending")
	ConditionTypeInProgress ConditionType = ConditionType("InProgress")
	ConditionTypeInSync     ConditionType = ConditionType("InSync")
)

type ConditionReason string

const (
	ConditionReasonPending    ConditionReason = ConditionReason("Pending")
	ConditionReasonInProgress ConditionReason = ConditionReason("InProgress")
	ConditionReasonInSync     ConditionReason = ConditionReason("InSync")
)
