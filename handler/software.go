package handler

import (
	"softhub/pkg/response"
	"softhub/request"
	"softhub/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SoftwareHandler struct {
	Service *service.SoftwareService
}

func NewSoftwareHandler(s *service.SoftwareService) *SoftwareHandler {
	return &SoftwareHandler{
		Service: s,
	}
}

// ---------------- 工具函数
func parseID(c *gin.Context, name string, label string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, label+"不合法")
		return 0, false
	}
	return id, true
}

// ----------------

// GET /software

func (h *SoftwareHandler) List(c *gin.Context) {
	softwares, err := h.Service.List(c.Request.Context())
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, softwares)
}

// POST /software

func (h *SoftwareHandler) Create(c *gin.Context) {
	var req []request.SoftwareReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不合法:"+err.Error())
		return
	}
	softwares, err := h.Service.CreateBatch(c.Request.Context(), req)
	if err != nil {
		response.ServerError(c, "创建软件失败")
		return
	}
	response.Success(c, softwares)
}

// GET /software/:id

func (h *SoftwareHandler) GetByID(c *gin.Context) {
	ID, ok := parseID(c, "id", "软件ID")
	if !ok {
		return
	}
	software, err := h.Service.GetByID(c.Request.Context(), ID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, software)
}

// PUT /software/:id

func (h *SoftwareHandler) Update(c *gin.Context) {
	var req request.SoftwareReq
	ID, ok := parseID(c, "id", "软件ID")
	if !ok {
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数不合法: "+err.Error())
		return
	}
	software, err := h.Service.Update(c.Request.Context(), ID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, software)
}

// DELETE /software/:id

func (h *SoftwareHandler) Delete(c *gin.Context) {
	ID, ok := parseID(c, "id", "软件ID")
	if !ok {
		return
	}
	software, err := h.Service.Delete(c.Request.Context(), ID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, software)
}

func (h *SoftwareHandler) CreateVersion(c *gin.Context) {
	ID, ok := parseID(c, "id", "软件ID")
	if !ok {
		return
	}
	var req request.VersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "版本号和下载地址不能为空")
		return
	}
	version, err := h.Service.CreateVersion(c.Request.Context(), ID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, version)
}
func (h *SoftwareHandler) ListVersion(c *gin.Context) {
	softwareID, ok := parseID(c, "id", "软件ID")
	if !ok {
		return
	}
	versions, err := h.Service.ListAllVersions(c.Request.Context(), softwareID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, versions)
}

// GET /software/:id/versions/latest

func (h *SoftwareHandler) GetLatestVersion(c *gin.Context) {
	softwareID, ok := parseID(c, "id", "软件ID")
	if !ok {
		return
	}
	version, err := h.Service.GetLatestVersion(c.Request.Context(), softwareID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, version)
}

// POST /software/:id/versions/:versionID/upload

func (h *SoftwareHandler) UploadVersionFile(c *gin.Context) {
	softwareID, ok := parseID(c, "id", "软件ID")
	if !ok {
		return
	}
	versionID, ok := parseID(c, "versionID", "版本ID")
	if !ok {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件(表单字段名必须为 file)")
		return
	}
	version, err := h.Service.UploadVersionFile(c.Request.Context(), softwareID, versionID, file)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, version)
}

// GET /software/:id/versions/:versionID/download

func (h *SoftwareHandler) DownloadVersionFile(c *gin.Context) {
	softwareID, ok := parseID(c, "id", "软件ID")
	if !ok {
		return
	}
	versionID, ok := parseID(c, "versionID", "版本ID")
	if !ok {
		return
	}
	filepath, downloadName, err := h.Service.GetDownloadFilePath(c.Request.Context(), softwareID, versionID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	c.FileAttachment(filepath, downloadName)
}
