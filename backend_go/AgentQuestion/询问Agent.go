package agentquestion

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	sqlinit "backend_go/SQLinit"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FlexibleUint uint

func (u *FlexibleUint) UnmarshalJSON(data []byte) error {
	var numberValue uint64
	if err := json.Unmarshal(data, &numberValue); err == nil {
		*u = FlexibleUint(numberValue)
		return nil
	}

	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		trimmed := strings.TrimSpace(stringValue)
		if trimmed == "" {
			return nil
		}
		parsed, parseErr := strconv.ParseUint(trimmed, 10, 64)
		if parseErr != nil {
			return parseErr
		}
		*u = FlexibleUint(parsed)
		return nil
	}

	return errors.New("invalid user_id")
}

type AskRequest struct {
	UserID       FlexibleUint           `json:"user_id" binding:"required"`
	ChatID       string                 `json:"chat_id"`
	Mode         string                 `json:"mode"`
	UserInput    string                 `json:"user_input"`
	Input        string                 `json:"input"`
	ImageURL     string                 `json:"image_url"`
	Image        string                 `json:"image"`
	ExistingPlan map[string]interface{} `json:"existing_plan"`
}

type ReplaceHistoryPlanRequest struct {
	UserID   FlexibleUint `json:"user_id" binding:"required"`
	RecordID uint         `json:"record_id" binding:"required"`
}

type streamEvent struct {
	Type  string          `json:"type"`
	Step  string          `json:"step,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

type agentMemoryItem struct {
	Memory string `json:"memory"`
	Tag    string `json:"tag"`
}

type memoryPipelineResponse struct {
	Summary  string            `json:"summary"`
	Memories []agentMemoryItem `json:"memories"`
}

func respondAgentError(c *gin.Context, status int, point string, message string, err error) {
	if err != nil {
		log.Printf("[%s] %v", point, err)
	} else {
		log.Printf("[%s] %s", point, message)
	}
	c.JSON(status, gin.H{"error": message, "error_point": point})
}

func resolveChatID(chatID string) string {
	trimmed := strings.TrimSpace(chatID)
	if trimmed != "" {
		return trimmed
	}
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func shouldAttachMemories(mode string) bool {
	return mode != "1_FoodRecognition"
}

func buildAgentUserProfile(mode string, profile sqlinit.UserProfile, memories []map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"height_cm":             profile.HeightCm,
		"weight_kg":             profile.WeightKg,
		"age":                   profile.Age,
		"gender":                profile.Gender,
		"core_goal":             profile.CoreGoal,
		"detailed_goal":         profile.DetailedGoal,
		"injury_history":        profile.InjuryHistory,
		"favorite_sports":       profile.FavoriteSports,
		"add_favorite_to_plan":  profile.AddFavoriteToPlan,
		"current_exercise_desc": profile.CurrentExerciseDesc,
		"expected_intensity":    profile.ExpectedIntensity,
	}
	if mode != "1_FoodRecognition" {
		result["memories"] = memories
	}
	if mode == "3_ChangePlan" {
		result["latest_plan_json"] = profile.LatestPlanJSON
	}
	return result
}

func normalizeMemoryTag(raw string) string {
	tag := strings.ToLower(strings.TrimSpace(raw))
	switch tag {
	case "constraint", "injury", "goal", "preference", "schedule", "device", "other":
		return tag
	default:
		return "other"
	}
}

func loadUserMemoriesForAgent(userID uint) ([]map[string]interface{}, error) {
	var records []sqlinit.UserMemoryRecord
	if err := sqlinit.DB.Where("user_id = ?", userID).Order("updated_at asc, id asc").Find(&records).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		memoryText := strings.TrimSpace(record.Memory)
		if memoryText == "" {
			continue
		}
		out = append(out, map[string]interface{}{
			"memory": memoryText,
			"tag":    normalizeMemoryTag(record.Tag),
		})
	}
	return out, nil
}

func replaceUserMemories(userID uint, items []agentMemoryItem) error {
	if len(items) == 0 {
		return nil
	}

	normalized := make([]agentMemoryItem, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		memoryText := strings.TrimSpace(item.Memory)
		if memoryText == "" {
			continue
		}
		tag := normalizeMemoryTag(item.Tag)
		key := tag + ":" + strings.ToLower(memoryText)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, agentMemoryItem{Memory: memoryText, Tag: tag})
	}
	if len(normalized) == 0 {
		return nil
	}

	return sqlinit.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&sqlinit.UserMemoryRecord{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, item := range normalized {
			record := sqlinit.UserMemoryRecord{
				UserID:    userID,
				Memory:    item.Memory,
				Tag:       item.Tag,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func syncUserMemoryFromDialog(userID uint, userInput string, profile sqlinit.UserProfile, memories []map[string]interface{}) error {
	trimmedInput := strings.TrimSpace(userInput)
	if trimmedInput == "" {
		return nil
	}

	payload := map[string]interface{}{
		"mode":       "5_Memory",
		"user_input": trimmedInput,
		"user_profile": buildAgentUserProfile(
			"5_Memory",
			profile,
			memories,
		),
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	agentURL := os.Getenv("AGENT_URL")
	if agentURL == "" {
		agentURL = "http://127.0.0.1:8000/pipeline/run"
	}
	req, err := http.NewRequest(http.MethodPost, agentURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return errors.New("memory pipeline failed")
	}

	var memoryResult memoryPipelineResponse
	if err := json.Unmarshal(respBytes, &memoryResult); err != nil {
		return err
	}
	if len(memoryResult.Memories) == 0 {
		return nil
	}
	if err := replaceUserMemories(userID, memoryResult.Memories); err != nil {
		return err
	}
	log.Printf("[ASK][S_MEMORY_SYNC] user_id=%d summary=%s memories=%d", userID, memoryResult.Summary, len(memoryResult.Memories))
	return nil
}

func cloneMemories(memories []map[string]interface{}) []map[string]interface{} {
	if len(memories) == 0 {
		return nil
	}
	cloned := make([]map[string]interface{}, 0, len(memories))
	for _, item := range memories {
		if item == nil {
			continue
		}
		copied := make(map[string]interface{}, len(item))
		for key, value := range item {
			copied[key] = value
		}
		cloned = append(cloned, copied)
	}
	return cloned
}

func syncUserMemoryFromDialogAsync(userID uint, userInput string, profile sqlinit.UserProfile, memories []map[string]interface{}, source string) {
	trimmedInput := strings.TrimSpace(userInput)
	if trimmedInput == "" {
		return
	}
	profileCopy := profile
	memoriesCopy := cloneMemories(memories)
	log.Printf("[%s][S_MEMORY_ASYNC_START] user_id=%d", source, userID)
	go func(input string, copiedProfile sqlinit.UserProfile, copiedMemories []map[string]interface{}) {
		startedAt := time.Now()
		if err := syncUserMemoryFromDialog(userID, input, copiedProfile, copiedMemories); err != nil {
			log.Printf("[%s][E_MEMORY_ASYNC] user_id=%d err=%v", source, userID, err)
			return
		}
		log.Printf("[%s][S_MEMORY_ASYNC_DONE] user_id=%d elapsed_ms=%d", source, userID, time.Since(startedAt).Milliseconds())
	}(trimmedInput, profileCopy, memoriesCopy)
}

func decodeImageDataURL(imageDataURL string) ([]byte, error) {
	trimmed := strings.TrimSpace(imageDataURL)
	if trimmed == "" {
		return nil, nil
	}
	if idx := strings.Index(trimmed, ","); idx >= 0 {
		trimmed = trimmed[idx+1:]
	}
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func buildAssistantRecord(mode string, result map[string]interface{}) (string, string, string, error) {
	if len(result) == 0 {
		return "", "", "", nil
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", "", "", err
	}

	if mode == "1_FoodRecognition" {
		foodName, _ := result["food_name_zh"].(string)
		if foodName == "" {
			foodName, _ = result["food_name_en"].(string)
		}
		if foodName == "" {
			foodName, _ = result["food_name"].(string) // 兼容旧版
		}
		if foodName != "" {
			return "识别结果: " + foodName, string(resultBytes), "food", nil
		}
		return string(resultBytes), string(resultBytes), "food", nil
	}

	if mode == "2_MakePlan" || mode == "3_ChangePlan" {
		if newPlan, ok := result["new_plan"]; ok {
			planBytes, marshalErr := json.Marshal(newPlan)
			if marshalErr != nil {
				return "", "", "", marshalErr
			}
			if mode == "2_MakePlan" {
				return "计划制定完成:", string(planBytes), "plan", nil
			}
			changeLines := extractChangeLines(result)
			if len(changeLines) == 0 {
				return "计划修改完成。", string(planBytes), "plan", nil
			}
			return "计划修改完成：\n- " + strings.Join(changeLines, "\n- "), string(planBytes), "plan", nil
		}
		if _, ok := result["months"]; ok {
			return "计划制定完成:", string(resultBytes), "plan", nil
		}
	}

	if mode == "4_OtherQuestion" {
		if answer, ok := result["answer"].(string); ok && strings.TrimSpace(answer) != "" {
			return strings.TrimSpace(answer), "", "", nil
		}
	}

	return string(resultBytes), "", "", nil
}

func extractChangeLines(result map[string]interface{}) []string {
	rawChanges, ok := result["changes"]
	if !ok {
		return nil
	}
	changeArray, ok := rawChanges.([]interface{})
	if !ok {
		return nil
	}
	lines := make([]string, 0, len(changeArray))
	for _, item := range changeArray {
		text := ""
		switch value := item.(type) {
		case string:
			text = value
		default:
			bytes, err := json.Marshal(value)
			if err == nil {
				text = string(bytes)
			}
		}
		text = strings.TrimSpace(text)
		if text != "" {
			lines = append(lines, text)
		}
	}
	return lines
}

func saveChatPair(userID uint, chatID string, mode string, userInput string, imageDataURL string, result map[string]interface{}, fallbackAssistantContent string) error {
	imageBytes, err := decodeImageDataURL(imageDataURL)
	if err != nil {
		return err
	}

	now := time.Now()
	userRecord := sqlinit.ChatMessageRecord{
		UserID:    userID,
		ChatID:    chatID,
		Role:      "user",
		Content:   userInput,
		ImageData: imageBytes,
		CreatedAt: now,
	}
	if err := sqlinit.DB.Create(&userRecord).Error; err != nil {
		return err
	}

	assistantContent, assistantJSONData, assistantJSONType, err := buildAssistantRecord(mode, result)
	if err != nil {
		return err
	}
	if assistantContent == "" && strings.TrimSpace(fallbackAssistantContent) != "" {
		assistantContent = strings.TrimSpace(fallbackAssistantContent)
	}
	if assistantContent == "" && assistantJSONData == "" {
		return nil
	}

	assistantRecord := sqlinit.ChatMessageRecord{
		UserID:    userID,
		ChatID:    chatID,
		Role:      "assistant",
		Content:   assistantContent,
		JSONData:  assistantJSONData,
		JSONType:  assistantJSONType,
		CreatedAt: now.Add(time.Millisecond),
	}
	return sqlinit.DB.Create(&assistantRecord).Error
}

func parsePlanJSON(planText string) interface{} {
	trimmed := strings.TrimSpace(planText)
	if trimmed == "" {
		return nil
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return planText
	}
	return parsed
}

func trimHistoryPlans(tx *gorm.DB, userID uint, keep int) error {
	if keep < 0 {
		keep = 0
	}
	var total int64
	if err := tx.Model(&sqlinit.PlanRecord{}).Where("user_id = ? AND plan_type = ?", userID, "history").Count(&total).Error; err != nil {
		return err
	}
	if total <= int64(keep) {
		return nil
	}

	staleCount := int(total) - keep
	var staleRecords []sqlinit.PlanRecord
	if err := tx.Where("user_id = ? AND plan_type = ?", userID, "history").Order("created_at desc, id desc").Offset(keep).Limit(staleCount).Find(&staleRecords).Error; err != nil {
		return err
	}
	if len(staleRecords) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(staleRecords))
	for _, item := range staleRecords {
		ids = append(ids, item.ID)
	}
	return tx.Where("id IN ?", ids).Delete(&sqlinit.PlanRecord{}).Error
}

func archiveCurrentPlanToHistory(tx *gorm.DB, userID uint, currentPlanJSON string) error {
	trimmed := strings.TrimSpace(currentPlanJSON)
	if trimmed == "" {
		return nil
	}
	var latestHistory sqlinit.PlanRecord
	err := tx.Where("user_id = ? AND plan_type = ?", userID, "history").Order("id desc").First(&latestHistory).Error
	if err == nil && strings.TrimSpace(latestHistory.PlanJSON) == trimmed {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	record := sqlinit.PlanRecord{
		UserID:   userID,
		PlanType: "history",
		PlanJSON: trimmed,
	}
	if err := tx.Create(&record).Error; err != nil {
		return err
	}
	return trimHistoryPlans(tx, userID, 2)
}

func saveLatestPlanWithHistory(tx *gorm.DB, userID uint, plan map[string]interface{}) error {
	normalizedPlan := plan
	if rawNewPlan, ok := plan["new_plan"]; ok {
		if casted, castOK := rawNewPlan.(map[string]interface{}); castOK {
			normalizedPlan = casted
		}
	}

	planJSONBytes, err := json.Marshal(normalizedPlan)
	if err != nil {
		return err
	}
	planJSON := string(planJSONBytes)

	var profile sqlinit.UserProfile
	err = tx.Select("user_id", "latest_plan_json").Where("user_id = ?", userID).First(&profile).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil && strings.TrimSpace(profile.LatestPlanJSON) != "" && strings.TrimSpace(profile.LatestPlanJSON) != strings.TrimSpace(planJSON) {
		if archiveErr := archiveCurrentPlanToHistory(tx, userID, profile.LatestPlanJSON); archiveErr != nil {
			return archiveErr
		}
	}

	row := sqlinit.UserProfile{UserID: userID}
	if err := tx.Where("user_id = ?", userID).FirstOrCreate(&row).Error; err != nil {
		return err
	}
	return tx.Model(&sqlinit.UserProfile{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"latest_plan_json":         planJSON,
		"latest_plan_generated_at": time.Now(),
	}).Error
}

func AskAgent(c *gin.Context) {
	log.Println("[ASK][S1] bind request")
	var request AskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondAgentError(c, http.StatusBadRequest, "E_ASK_BIND", "参数错误", err)
		return
	}
	if uint(request.UserID) == 0 {
		respondAgentError(c, http.StatusBadRequest, "E_ASK_USER_ID", "user_id无效", nil)
		return
	}

	if strings.TrimSpace(request.UserInput) == "" {
		request.UserInput = strings.TrimSpace(request.Input)
	}
	if strings.TrimSpace(request.ImageURL) == "" {
		request.ImageURL = strings.TrimSpace(request.Image)
	}
	if request.UserInput == "" && request.ImageURL == "" {
		respondAgentError(c, http.StatusBadRequest, "E_ASK_INPUT_EMPTY", "user_input和image至少传一个", nil)
		return
	}
	request.Mode = normalizeMode(request.Mode)
	chatID := resolveChatID(request.ChatID)
	if request.Mode == "4_OtherQuestion" && strings.TrimSpace(request.UserInput) == "" {
		respondAgentError(c, http.StatusBadRequest, "E_ASK_MODE4_TEXT_REQUIRED", "其他模式必须提供文字问题", nil)
		return
	}
	userID := uint(request.UserID)
	log.Printf(
		"[AskAgent] start user_id=%d mode=%s text_len=%d has_image=%t",
		userID,
		request.Mode,
		len(request.UserInput),
		request.ImageURL != "",
	)

	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ASK][S3] profile not found, continue with empty profile user_id=%d", userID)
		} else {
			respondAgentError(c, http.StatusInternalServerError, "E_ASK_PROFILE_QUERY", "查询用户信息失败", err)
			return
		}
	} else {
		log.Printf("[ASK][S3] profile loaded user_id=%d", userID)
	}

	var userMemories []map[string]interface{}
	if shouldAttachMemories(request.Mode) {
		loadedMemories, memoryErr := loadUserMemoriesForAgent(userID)
		if memoryErr != nil {
			respondAgentError(c, http.StatusInternalServerError, "E_ASK_MEMORY_QUERY", "查询用户记忆失败", memoryErr)
			return
		}
		userMemories = loadedMemories
	}

	payload := map[string]interface{}{
		"mode":          request.Mode,
		"user_input":    request.UserInput,
		"image_url":     request.ImageURL,
		"existing_plan": request.ExistingPlan,
		"user_profile":  buildAgentUserProfile(request.Mode, profile, userMemories),
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_ASK_PAYLOAD_MARSHAL", "请求构造失败", err)
		return
	}
	agentURL := os.Getenv("AGENT_URL")
	if agentURL == "" {
		agentURL = "http://127.0.0.1:8000/pipeline/run"
	}
	log.Printf("[AskAgent] forwarding to agent url=%s payload_bytes=%d", agentURL, len(bodyBytes))

	req, err := http.NewRequest(http.MethodPost, agentURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_ASK_HTTP_NEW_REQUEST", "构造agent请求失败", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 900 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		respondAgentError(c, http.StatusBadGateway, "E_ASK_AGENT_UNAVAILABLE", "agent服务不可用", err)
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		respondAgentError(c, http.StatusBadGateway, "E_ASK_AGENT_READ", "读取agent响应失败", err)
		return
	}
	preview := string(respBytes)
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	log.Printf("[AskAgent] agent status=%d response_preview=%s", resp.StatusCode, preview)
	var agentResult map[string]interface{}
	assistantFallback := ""
	if err := json.Unmarshal(respBytes, &agentResult); err != nil {
		log.Printf("[ASK][S4] agent response is not json, skip plan persistence: %v", err)
		assistantFallback = strings.TrimSpace(string(respBytes))
	} else if request.Mode == "4_OtherQuestion" {
		assistantFallback = strings.TrimSpace(string(respBytes))
	}

	if (request.Mode == "2_MakePlan" || request.Mode == "3_ChangePlan") && len(agentResult) > 0 {
		if err := saveLatestPlanWithHistory(sqlinit.DB, userID, agentResult); err != nil {
			respondAgentError(c, http.StatusInternalServerError, "E_ASK_PROFILE_PLAN_UPDATE", "更新用户计划失败", err)
			return
		}
		log.Printf("[ASK][S5] latest plan updated with history user_id=%d mode=%s", userID, request.Mode)
	}

	if resp.StatusCode < http.StatusBadRequest {
		if err := saveChatPair(userID, chatID, request.Mode, request.UserInput, request.ImageURL, agentResult, assistantFallback); err != nil {
			log.Printf("[ASK][E_ASK_CHAT_RECORD_CREATE] save chat record failed user_id=%d chat_id=%s err=%v", userID, chatID, err)
		}
		if request.Mode != "1_FoodRecognition" {
			syncUserMemoryFromDialogAsync(userID, request.UserInput, profile, userMemories, "ASK")
		}
	}

	c.Data(resp.StatusCode, "application/json", respBytes)
}

func AskAgentStream(c *gin.Context) {
	log.Println("[ASK_STREAM][S1] bind request")
	var request AskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondAgentError(c, http.StatusBadRequest, "E_ASK_STREAM_BIND", "参数错误", err)
		return
	}
	if uint(request.UserID) == 0 {
		respondAgentError(c, http.StatusBadRequest, "E_ASK_STREAM_USER_ID", "user_id无效", nil)
		return
	}

	if strings.TrimSpace(request.UserInput) == "" {
		request.UserInput = strings.TrimSpace(request.Input)
	}
	if strings.TrimSpace(request.ImageURL) == "" {
		request.ImageURL = strings.TrimSpace(request.Image)
	}
	if request.UserInput == "" && request.ImageURL == "" {
		respondAgentError(c, http.StatusBadRequest, "E_ASK_STREAM_INPUT_EMPTY", "user_input和image至少传一个", nil)
		return
	}

	request.Mode = normalizeMode(request.Mode)
	chatID := resolveChatID(request.ChatID)
	if request.Mode == "4_OtherQuestion" && strings.TrimSpace(request.UserInput) == "" {
		respondAgentError(c, http.StatusBadRequest, "E_ASK_STREAM_MODE4_TEXT_REQUIRED", "其他模式必须提供文字问题", nil)
		return
	}
	userID := uint(request.UserID)
	log.Printf(
		"[AskAgentStream] start user_id=%d mode=%s text_len=%d has_image=%t",
		userID,
		request.Mode,
		len(request.UserInput),
		request.ImageURL != "",
	)

	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ASK_STREAM][S2] profile not found, continue with empty profile user_id=%d", userID)
		} else {
			respondAgentError(c, http.StatusInternalServerError, "E_ASK_STREAM_PROFILE_QUERY", "查询用户信息失败", err)
			return
		}
	} else {
		log.Printf("[ASK_STREAM][S2] profile loaded user_id=%d", userID)
	}

	var userMemories []map[string]interface{}
	if shouldAttachMemories(request.Mode) {
		loadedMemories, memoryErr := loadUserMemoriesForAgent(userID)
		if memoryErr != nil {
			respondAgentError(c, http.StatusInternalServerError, "E_ASK_STREAM_MEMORY_QUERY", "查询用户记忆失败", memoryErr)
			return
		}
		userMemories = loadedMemories
	}

	payload := map[string]interface{}{
		"mode":          request.Mode,
		"user_input":    request.UserInput,
		"image_url":     request.ImageURL,
		"existing_plan": request.ExistingPlan,
		"user_profile":  buildAgentUserProfile(request.Mode, profile, userMemories),
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_ASK_STREAM_PAYLOAD_MARSHAL", "请求构造失败", err)
		return
	}
	agentURL := os.Getenv("AGENT_STREAM_URL")
	if agentURL == "" {
		baseAgentURL := os.Getenv("AGENT_URL")
		if baseAgentURL == "" {
			agentURL = "http://127.0.0.1:8000/pipeline/run_stream"
		} else {
			agentURL = strings.Replace(baseAgentURL, "/pipeline/run", "/pipeline/run_stream", 1)
		}
	}
	log.Printf("[AskAgentStream] forwarding to agent url=%s payload_bytes=%d", agentURL, len(bodyBytes))

	req, err := http.NewRequest(http.MethodPost, agentURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_ASK_STREAM_HTTP_NEW_REQUEST", "构造agent请求失败", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 900 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		respondAgentError(c, http.StatusBadGateway, "E_ASK_STREAM_AGENT_UNAVAILABLE", "agent服务不可用", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			respondAgentError(c, http.StatusBadGateway, "E_ASK_STREAM_AGENT_READ_ERROR_BODY", "读取agent错误响应失败", err)
			return
		}
		c.Data(resp.StatusCode, "application/json", respBytes)
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		respondAgentError(c, http.StatusInternalServerError, "E_ASK_STREAM_FLUSHER_UNSUPPORTED", "stream不支持", nil)
		return
	}

	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	scanner := bufio.NewScanner(resp.Body)
	buffer := make([]byte, 0, 1024*64)
	scanner.Buffer(buffer, 1024*1024*8)

	var finalResult map[string]interface{}
	finalAssistantFallback := ""
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event streamEvent
		if err := json.Unmarshal(line, &event); err == nil {
			dataPreview := ""
			if len(event.Data) > 0 {
				dataPreview = string(event.Data)
				if len(dataPreview) > 500 {
					dataPreview = dataPreview[:500] + "..."
				}
			}
			log.Printf(
				"[AskAgentStream] event user_id=%d mode=%s type=%s step=%s error=%s data_preview=%s",
				userID,
				request.Mode,
				event.Type,
				event.Step,
				event.Error,
				dataPreview,
			)
			if event.Type == "result" {
				if err := json.Unmarshal(event.Data, &finalResult); err != nil {
					log.Printf("[ASK_STREAM][S3] result parse failed user_id=%d mode=%s err=%v", userID, request.Mode, err)
					var textResult string
					if textErr := json.Unmarshal(event.Data, &textResult); textErr == nil {
						finalAssistantFallback = strings.TrimSpace(textResult)
					}
				} else if request.Mode == "4_OtherQuestion" {
					finalAssistantFallback = strings.TrimSpace(string(event.Data))
				}
			}
		} else {
			log.Printf("[ASK_STREAM][S3] ignore invalid ndjson line user_id=%d mode=%s err=%v", userID, request.Mode, err)
		}

		if _, err := c.Writer.Write(append(append([]byte{}, line...), '\n')); err != nil {
			log.Printf("[ASK_STREAM][E_ASK_STREAM_WRITE] write stream failed user_id=%d mode=%s err=%v", userID, request.Mode, err)
			break
		}
		flusher.Flush()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[AskAgentStream] scanner error user_id=%d mode=%s err=%v", userID, request.Mode, err)
	}

	if (request.Mode == "2_MakePlan" || request.Mode == "3_ChangePlan") && len(finalResult) > 0 {
		if err := saveLatestPlanWithHistory(sqlinit.DB, userID, finalResult); err != nil {
			log.Printf("[ASK_STREAM][E_ASK_STREAM_PROFILE_PLAN_UPDATE] update profile plan failed user_id=%d mode=%s err=%v", userID, request.Mode, err)
			return
		}
		log.Printf("[ASK_STREAM][S4] latest plan updated with history user_id=%d mode=%s", userID, request.Mode)
	}

	if len(finalResult) > 0 || finalAssistantFallback != "" {
		if err := saveChatPair(userID, chatID, request.Mode, request.UserInput, request.ImageURL, finalResult, finalAssistantFallback); err != nil {
			log.Printf("[ASK_STREAM][E_ASK_STREAM_CHAT_RECORD_CREATE] save chat record failed user_id=%d chat_id=%s err=%v", userID, chatID, err)
		}
		if request.Mode != "1_FoodRecognition" {
			syncUserMemoryFromDialogAsync(userID, request.UserInput, profile, userMemories, "ASK_STREAM")
		}
	}
}

func GetChatHistory(c *gin.Context) {
	log.Println("[CHAT_HISTORY][S1] parse query")
	userIDValue, err := strconv.ParseUint(strings.TrimSpace(c.Query("user_id")), 10, 64)
	if err != nil || userIDValue == 0 {
		respondAgentError(c, http.StatusBadRequest, "E_CHAT_HISTORY_USER_ID", "user_id无效", err)
		return
	}
	chatID := strings.TrimSpace(c.Query("chat_id"))
	if chatID == "" {
		respondAgentError(c, http.StatusBadRequest, "E_CHAT_HISTORY_CHAT_ID", "chat_id不能为空", nil)
		return
	}

	pairLimit := 5
	if rawPairLimit := strings.TrimSpace(c.Query("pair_limit")); rawPairLimit != "" {
		parsed, parseErr := strconv.Atoi(rawPairLimit)
		if parseErr != nil || parsed <= 0 {
			respondAgentError(c, http.StatusBadRequest, "E_CHAT_HISTORY_PAIR_LIMIT", "pair_limit无效", parseErr)
			return
		}
		if parsed > 20 {
			parsed = 20
		}
		pairLimit = parsed
	}

	beforeID := uint64(0)
	if rawBeforeID := strings.TrimSpace(c.Query("before_id")); rawBeforeID != "" {
		parsed, parseErr := strconv.ParseUint(rawBeforeID, 10, 64)
		if parseErr != nil {
			respondAgentError(c, http.StatusBadRequest, "E_CHAT_HISTORY_BEFORE_ID", "before_id无效", parseErr)
			return
		}
		beforeID = parsed
	}

	messageLimit := pairLimit * 2
	query := sqlinit.DB.Where("user_id = ? AND chat_id = ?", uint(userIDValue), chatID)
	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}

	var records []sqlinit.ChatMessageRecord
	if err := query.Order("id desc").Limit(messageLimit + 1).Find(&records).Error; err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_CHAT_HISTORY_DB_QUERY", "查询聊天记录失败", err)
		return
	}

	hasMore := len(records) > messageLimit
	if hasMore {
		records = records[:messageLimit]
	}

	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}

	messages := make([]gin.H, 0, len(records))
	for _, record := range records {
		imageBase64 := ""
		if len(record.ImageData) > 0 {
			imageBase64 = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(record.ImageData)
		}
		messages = append(messages, gin.H{
			"id":           record.ID,
			"role":         record.Role,
			"content":      record.Content,
			"image_base64": imageBase64,
			"json_data":    record.JSONData,
			"json_type":    record.JSONType,
			"created_at":   record.CreatedAt,
		})
	}

	var nextBeforeID interface{} = nil
	if hasMore && len(records) > 0 {
		nextBeforeID = records[0].ID
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":       messages,
		"has_more":       hasMore,
		"next_before_id": nextBeforeID,
	})
}

func DeleteChatHistory(c *gin.Context) {
	log.Println("[CHAT_HISTORY_DELETE][S1] parse query")
	userIDValue, err := strconv.ParseUint(strings.TrimSpace(c.Query("user_id")), 10, 64)
	if err != nil || userIDValue == 0 {
		respondAgentError(c, http.StatusBadRequest, "E_CHAT_HISTORY_DELETE_USER_ID", "user_id无效", err)
		return
	}
	chatID := strings.TrimSpace(c.Query("chat_id"))
	if chatID == "" {
		respondAgentError(c, http.StatusBadRequest, "E_CHAT_HISTORY_DELETE_CHAT_ID", "chat_id不能为空", nil)
		return
	}

	if err := sqlinit.DB.Where("user_id = ? AND chat_id = ?", uint(userIDValue), chatID).Delete(&sqlinit.ChatMessageRecord{}).Error; err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_CHAT_HISTORY_DELETE_DB", "删除聊天记录失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetPlanHistory(c *gin.Context) {
	log.Println("[PLAN_HISTORY][S1] parse query")
	userIDValue, err := strconv.ParseUint(strings.TrimSpace(c.Query("user_id")), 10, 64)
	if err != nil || userIDValue == 0 {
		respondAgentError(c, http.StatusBadRequest, "E_PLAN_HISTORY_USER_ID", "user_id无效", err)
		return
	}

	userID := uint(userIDValue)
	var records []sqlinit.PlanRecord
	if err := sqlinit.DB.Where("user_id = ? AND plan_type = ?", userID, "history").Order("created_at desc, id desc").Limit(2).Find(&records).Error; err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_PLAN_HISTORY_QUERY", "查询历史计划失败", err)
		return
	}

	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		items = append(items, gin.H{
			"id":         record.ID,
			"plan_type":  record.PlanType,
			"plan":       parsePlanJSON(record.PlanJSON),
			"created_at": record.CreatedAt,
		})
	}

	currentPlan := interface{}(nil)
	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Select("latest_plan_json").Where("user_id = ?", userID).First(&profile).Error; err == nil {
		currentPlan = parsePlanJSON(profile.LatestPlanJSON)
	}

	c.JSON(http.StatusOK, gin.H{
		"current_plan": currentPlan,
		"items":        items,
	})
}

func DeletePlanHistory(c *gin.Context) {
	log.Println("[PLAN_HISTORY_DELETE][S1] parse query")
	userIDValue, err := strconv.ParseUint(strings.TrimSpace(c.Query("user_id")), 10, 64)
	if err != nil || userIDValue == 0 {
		respondAgentError(c, http.StatusBadRequest, "E_PLAN_HISTORY_DELETE_USER_ID", "user_id无效", err)
		return
	}
	recordIDValue, err := strconv.ParseUint(strings.TrimSpace(c.Query("record_id")), 10, 64)
	if err != nil || recordIDValue == 0 {
		respondAgentError(c, http.StatusBadRequest, "E_PLAN_HISTORY_DELETE_RECORD_ID", "record_id无效", err)
		return
	}

	result := sqlinit.DB.Where("id = ? AND user_id = ? AND plan_type = ?", uint(recordIDValue), uint(userIDValue), "history").Delete(&sqlinit.PlanRecord{})
	if result.Error != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_PLAN_HISTORY_DELETE_DB", "删除历史计划失败", result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondAgentError(c, http.StatusNotFound, "E_PLAN_HISTORY_DELETE_NOT_FOUND", "历史计划不存在", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ReplacePlanFromHistory(c *gin.Context) {
	log.Println("[PLAN_HISTORY_REPLACE][S1] bind request")
	var request ReplaceHistoryPlanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondAgentError(c, http.StatusBadRequest, "E_PLAN_HISTORY_REPLACE_BIND", "参数错误", err)
		return
	}
	if uint(request.UserID) == 0 || request.RecordID == 0 {
		respondAgentError(c, http.StatusBadRequest, "E_PLAN_HISTORY_REPLACE_REQUIRED", "user_id和record_id必填", nil)
		return
	}

	userID := uint(request.UserID)
	var selectedRecord sqlinit.PlanRecord
	if err := sqlinit.DB.Where("id = ? AND user_id = ? AND plan_type = ?", request.RecordID, userID, "history").First(&selectedRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondAgentError(c, http.StatusNotFound, "E_PLAN_HISTORY_REPLACE_NOT_FOUND", "历史计划不存在", err)
			return
		}
		respondAgentError(c, http.StatusInternalServerError, "E_PLAN_HISTORY_REPLACE_QUERY", "读取历史计划失败", err)
		return
	}

	err := sqlinit.DB.Transaction(func(tx *gorm.DB) error {
		var profile sqlinit.UserProfile
		profileErr := tx.Select("user_id", "latest_plan_json").Where("user_id = ?", userID).First(&profile).Error
		if profileErr != nil && !errors.Is(profileErr, gorm.ErrRecordNotFound) {
			return profileErr
		}
		currentPlanJSON := ""
		if profileErr == nil {
			currentPlanJSON = strings.TrimSpace(profile.LatestPlanJSON)
		}
		selectedPlanJSON := strings.TrimSpace(selectedRecord.PlanJSON)
		isSamePlan := currentPlanJSON != "" && currentPlanJSON == selectedPlanJSON

		if !isSamePlan && currentPlanJSON != "" {
			if archiveErr := archiveCurrentPlanToHistory(tx, userID, currentPlanJSON); archiveErr != nil {
				return archiveErr
			}
		}

		row := sqlinit.UserProfile{UserID: userID}
		if firstErr := tx.Where("user_id = ?", userID).FirstOrCreate(&row).Error; firstErr != nil {
			return firstErr
		}
		if updateErr := tx.Model(&sqlinit.UserProfile{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
			"latest_plan_json":         selectedRecord.PlanJSON,
			"latest_plan_generated_at": time.Now(),
		}).Error; updateErr != nil {
			return updateErr
		}

		if !isSamePlan {
			if delErr := tx.Where("id = ? AND user_id = ? AND plan_type = ?", selectedRecord.ID, userID, "history").Delete(&sqlinit.PlanRecord{}).Error; delErr != nil {
				return delErr
			}
		}

		return trimHistoryPlans(tx, userID, 2)
	})
	if err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_PLAN_HISTORY_REPLACE_TX", "替换当前计划失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":           true,
		"current_plan": parsePlanJSON(selectedRecord.PlanJSON),
	})
}

func normalizeMode(mode string) string {
	trimmed := strings.TrimSpace(mode)
	switch trimmed {
	case "1", "1_FoodRecognition":
		return "1_FoodRecognition"
	case "2", "2_MakePlan":
		return "2_MakePlan"
	case "3", "3_ChangePlan":
		return "3_ChangePlan"
	default:
		return "4_OtherQuestion"
	}
}
