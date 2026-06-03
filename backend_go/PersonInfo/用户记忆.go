package personinfo

import (
	"net/http"
	"strconv"
	"strings"

	sqlinit "backend_go/SQLinit"

	"github.com/gin-gonic/gin"
)

func GetMemoryList(c *gin.Context) {
	userIDValue, err := strconv.ParseUint(strings.TrimSpace(c.Query("user_id")), 10, 64)
	if err != nil || userIDValue == 0 {
		respondError(c, http.StatusBadRequest, "E_MEMORY_LIST_USER_ID", "user_id无效", err)
		return
	}

	var records []sqlinit.UserMemoryRecord
	if err := sqlinit.DB.Where("user_id = ?", uint(userIDValue)).Order("updated_at desc, id desc").Find(&records).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "E_MEMORY_LIST_QUERY", "查询记忆失败", err)
		return
	}

	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		items = append(items, gin.H{
			"id":         record.ID,
			"memory":     record.Memory,
			"tag":        record.Tag,
			"created_at": record.CreatedAt,
			"updated_at": record.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}

func DeleteMemory(c *gin.Context) {
	userIDValue, err := strconv.ParseUint(strings.TrimSpace(c.Query("user_id")), 10, 64)
	if err != nil || userIDValue == 0 {
		respondError(c, http.StatusBadRequest, "E_MEMORY_DELETE_USER_ID", "user_id无效", err)
		return
	}
	memoryIDValue, err := strconv.ParseUint(strings.TrimSpace(c.Query("memory_id")), 10, 64)
	if err != nil || memoryIDValue == 0 {
		respondError(c, http.StatusBadRequest, "E_MEMORY_DELETE_MEMORY_ID", "memory_id无效", err)
		return
	}

	result := sqlinit.DB.Where("id = ? AND user_id = ?", uint(memoryIDValue), uint(userIDValue)).Delete(&sqlinit.UserMemoryRecord{})
	if result.Error != nil {
		respondError(c, http.StatusInternalServerError, "E_MEMORY_DELETE_DB", "删除记忆失败", result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondError(c, http.StatusNotFound, "E_MEMORY_DELETE_NOT_FOUND", "记忆不存在", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
