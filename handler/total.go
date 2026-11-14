package handler

type TotalHandler struct {
	Software *SoftwareHandler
	User     *UserHandler
}

func NewTotalHandler(software *SoftwareHandler, user *UserHandler) *TotalHandler {
	return &TotalHandler{
		Software: software,
		User:     user,
	}
}
