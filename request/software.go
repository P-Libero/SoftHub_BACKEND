package request

type SoftwareReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Version     string `json:"version" binding:"required"`
}
