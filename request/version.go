package request

type VersionReq struct {
	Version     string `json:"version" binding:"required"`
	Description string `json:"description"`
}
