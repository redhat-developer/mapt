package amireplication

type ReplicatedRequest struct {
	ProjectName     string
	AMITargetName   string
	AMISourceID     string
	AMISourceRegion string
}
