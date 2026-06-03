package personinfo

import (
	"encoding/base64"
	"errors"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	sqlinit "backend_go/SQLinit"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SaveProfileRequest struct {
	UserID                 uint     `json:"user_id" binding:"required"`
	HeightCm               *float64 `json:"height_cm"`
	Height                 *float64 `json:"height"`
	WeightKg               *float64 `json:"weight_kg"`
	Weight                 *float64 `json:"weight"`
	WeightUnit             string   `json:"weight_unit"`
	BirthDate              string   `json:"birth_date"`
	Age                    *int     `json:"age"`
	Gender                 string   `json:"gender"`
	CoreGoal               string   `json:"core_goal"`
	CoreGoals              []string `json:"core_goals"`
	DetailedGoal           string   `json:"detailed_goal"`
	InjuryHistory          string   `json:"injury_history"`
	FavoriteSports         string   `json:"favorite_sports"`
	AddFavoriteToPlan      bool     `json:"add_favorite_to_plan"`
	AutoAddToPlan          *bool    `json:"auto_add_to_plan"`
	CurrentExerciseDesc    string   `json:"current_exercise_desc"`
	SportsDescription      string   `json:"sports_description"`
	ExpectedIntensity      string   `json:"expected_intensity"`
	TargetIntensity        string   `json:"target_intensity"`
	EnableFavoritePlanDesc *bool    `json:"enable_favorite_plan_desc"`
}

func SaveProfile(c *gin.Context) {
	log.Println("[PROFILE][SAVE][S1] bind request")
	var request SaveProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "E_PROFILE_SAVE_BIND", "参数错误", err)
		return
	}
	log.Printf("[PROFILE][SAVE][S2] validate required fields user_id=%d", request.UserID)

	height := pickFloat(request.HeightCm, request.Height)
	weight := pickFloat(request.WeightKg, request.Weight)
	weightUnit := normalizeWeightUnit(request.WeightUnit)
	if height <= 0 || weight <= 0 || weightUnit == "" {
		respondError(c, http.StatusBadRequest, "E_PROFILE_SAVE_HEIGHT_WEIGHT", "身高和体重必填", nil)
		return
	}
	weightKg := toKg(weight, weightUnit)
	if weightKg <= 0 {
		respondError(c, http.StatusBadRequest, "E_PROFILE_SAVE_WEIGHT_UNIT", "体重单位错误", nil)
		return
	}

	gender := strings.TrimSpace(request.Gender)
	if gender == "" {
		respondError(c, http.StatusBadRequest, "E_PROFILE_SAVE_GENDER", "性别必填", nil)
		return
	}

	expectedIntensity := strings.TrimSpace(request.ExpectedIntensity)
	if expectedIntensity == "" {
		expectedIntensity = strings.TrimSpace(request.TargetIntensity)
	}
	if expectedIntensity == "" {
		respondError(c, http.StatusBadRequest, "E_PROFILE_SAVE_INTENSITY", "训练强度必填", nil)
		return
	}

	coreGoals := normalizeCoreGoals(request.CoreGoal, request.CoreGoals)
	if len(coreGoals) == 0 {
		respondError(c, http.StatusBadRequest, "E_PROFILE_SAVE_CORE_GOAL", "核心诉求必填", nil)
		return
	}

	detailedGoal := strings.TrimSpace(request.DetailedGoal)
	if containsGoal(coreGoals, "其他") && detailedGoal == "" {
		respondError(c, http.StatusBadRequest, "E_PROFILE_SAVE_OTHER_NEED_DETAIL", "选择“其他”时必须填写详细诉求", nil)
		return
	}

	birthDateText := strings.TrimSpace(request.BirthDate)
	if birthDateText == "" {
		respondError(c, http.StatusBadRequest, "E_PROFILE_SAVE_BIRTH_DATE_EMPTY", "出生日期必填", nil)
		return
	}
	birthDate, err := time.Parse("2006-01-02", birthDateText)
	if err != nil {
		respondError(c, http.StatusBadRequest, "E_PROFILE_SAVE_BIRTH_DATE_FORMAT", "出生日期格式错误，应为YYYY-MM-DD", err)
		return
	}
	age := calcAge(birthDate)
	if request.Age != nil && *request.Age > 0 {
		age = *request.Age
	}

	addFavoriteToPlan := request.AddFavoriteToPlan
	if request.AutoAddToPlan != nil {
		addFavoriteToPlan = *request.AutoAddToPlan
	}
	if request.EnableFavoritePlanDesc != nil {
		addFavoriteToPlan = *request.EnableFavoritePlanDesc
	}

	currentExerciseDesc := strings.TrimSpace(request.CurrentExerciseDesc)
	if currentExerciseDesc == "" {
		currentExerciseDesc = strings.TrimSpace(request.SportsDescription)
	}
	if !addFavoriteToPlan {
		currentExerciseDesc = ""
	}
	log.Printf("[PROFILE][SAVE][S3] normalized payload user_id=%d core_goals=%d add_favorite=%t", request.UserID, len(coreGoals), addFavoriteToPlan)

	profile := sqlinit.UserProfile{
		UserID:                request.UserID,
		HeightCm:              height,
		WeightKg:              weightKg,
		WeightUnit:            weightUnit,
		WeightLocked:          false,
		Age:                   age,
		BirthDate:             &birthDate,
		Gender:                gender,
		CoreGoal:              strings.Join(coreGoals, ","),
		DetailedGoal:          detailedGoal,
		InjuryHistory:         strings.TrimSpace(request.InjuryHistory),
		FavoriteSports:        strings.TrimSpace(request.FavoriteSports),
		AddFavoriteToPlan:     addFavoriteToPlan,
		CurrentExerciseDesc:   currentExerciseDesc,
		ExpectedIntensity:     expectedIntensity,
		LatestPlanGeneratedAt: time.Now(),
	}

	var existing sqlinit.UserProfile
	existingErr := sqlinit.DB.Where("user_id = ?", request.UserID).First(&existing).Error
	if existingErr != nil && !errors.Is(existingErr, gorm.ErrRecordNotFound) {
		respondError(c, http.StatusInternalServerError, "E_PROFILE_SAVE_QUERY", "查询用户信息失败", existingErr)
		return
	}
	if errors.Is(existingErr, gorm.ErrRecordNotFound) {
		if err := sqlinit.DB.Create(&profile).Error; err != nil {
			respondError(c, http.StatusInternalServerError, "E_PROFILE_SAVE_DB_CREATE", "保存失败", err)
			return
		}
	} else {
		updates := map[string]interface{}{
			"height_cm":                profile.HeightCm,
			"weight_kg":                weightKg,
			"weight_unit":              weightUnit,
			"weight_locked":            false,
			"age":                      profile.Age,
			"birth_date":               profile.BirthDate,
			"gender":                   profile.Gender,
			"core_goal":                profile.CoreGoal,
			"detailed_goal":            profile.DetailedGoal,
			"injury_history":           profile.InjuryHistory,
			"favorite_sports":          profile.FavoriteSports,
			"add_favorite_to_plan":     profile.AddFavoriteToPlan,
			"current_exercise_desc":    profile.CurrentExerciseDesc,
			"expected_intensity":       profile.ExpectedIntensity,
			"latest_plan_generated_at": profile.LatestPlanGeneratedAt,
		}
		if err := sqlinit.DB.Model(&existing).Updates(updates).Error; err != nil {
			respondError(c, http.StatusInternalServerError, "E_PROFILE_SAVE_DB_UPDATE", "保存失败", err)
			return
		}
	}

	if err := ensureWeightParam(request.UserID, gender); err != nil {
		respondError(c, http.StatusInternalServerError, "E_PROFILE_SAVE_DB", "保存失败", err)
		return
	}
	log.Printf("[PROFILE][SAVE][S4] save success user_id=%d", request.UserID)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func GetProfile(c *gin.Context) {
	log.Println("[PROFILE][GET][S1] parse query user_id")
	userIDText := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDText, 10, 64)
	if err != nil || userID == 0 {
		respondError(c, http.StatusBadRequest, "E_PROFILE_GET_USER_ID", "user_id参数错误", err)
		return
	}
	log.Printf("[PROFILE][GET][S2] query user user_id=%d", userID)

	var user sqlinit.User
	if err := sqlinit.DB.First(&user, uint(userID)).Error; err != nil {
		respondError(c, http.StatusNotFound, "E_PROFILE_GET_USER_QUERY", "用户不存在", err)
		return
	}

	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[PROFILE][GET][S3] profile not found user_id=%d", userID)
			c.JSON(http.StatusOK, gin.H{
				"user_id":        userID,
				"username":       user.Username,
				"avatar_base64":  toAvatarBase64(user.AvatarData),
				"profile_exists": false,
			})
			return
		}
		respondError(c, http.StatusInternalServerError, "E_PROFILE_GET_PROFILE_QUERY", "查询用户信息失败", err)
		return
	}
	log.Printf("[PROFILE][GET][S4] profile loaded user_id=%d", userID)

	coreGoals := splitCoreGoals(profile.CoreGoal)
	weightUnit := normalizeWeightUnit(profile.WeightUnit)
	if weightUnit == "" {
		weightUnit = "kg"
	}
	resolveProfileWeightedKg(&profile)
	weightDisplay := fromKg(profile.WeightKg, weightUnit)
	bmiValue := calcBMI(profile.WeightKg, profile.HeightCm)
	bmiDisplay := interface{}(nil)
	bmiStatus := "未知"
	if bmiValue > 0 {
		bmiDisplay = roundOne(bmiValue)
		bmiStatus = bmiLevel(bmiValue)
	}
	if profile.WeightLocked {
		profile.WeightLocked = false
		_ = sqlinit.DB.Model(&profile).Updates(map[string]interface{}{"weight_locked": false, "weight_unit": weightUnit}).Error
	}
	c.JSON(http.StatusOK, gin.H{
		"id":                       profile.ID,
		"user_id":                  profile.UserID,
		"username":                 user.Username,
		"avatar_base64":            toAvatarBase64(user.AvatarData),
		"profile_exists":           true,
		"height_cm":                profile.HeightCm,
		"height":                   profile.HeightCm,
		"weight_kg":                profile.WeightKg,
		"weight":                   roundOne(weightDisplay),
		"weight_unit":              weightUnit,
		"bmi":                      bmiDisplay,
		"bmi_status":               bmiStatus,
		"weight_locked":            profile.WeightLocked,
		"age":                      profile.Age,
		"birth_date":               formatDatePtr(profile.BirthDate),
		"gender":                   profile.Gender,
		"core_goal":                profile.CoreGoal,
		"core_goals":               coreGoals,
		"detailed_goal":            profile.DetailedGoal,
		"injury_history":           profile.InjuryHistory,
		"favorite_sports":          profile.FavoriteSports,
		"add_favorite_to_plan":     profile.AddFavoriteToPlan,
		"auto_add_to_plan":         profile.AddFavoriteToPlan,
		"current_exercise_desc":    profile.CurrentExerciseDesc,
		"sports_description":       profile.CurrentExerciseDesc,
		"expected_intensity":       profile.ExpectedIntensity,
		"target_intensity":         profile.ExpectedIntensity,
		"latest_plan_json":         profile.LatestPlanJSON,
		"latest_plan_generated_at": profile.LatestPlanGeneratedAt,
	})
}

type SaveWeightRecordRequest struct {
	UserID     uint    `json:"user_id" binding:"required"`
	Weight     float64 `json:"weight" binding:"required"`
	WeightUnit string  `json:"weight_unit"`
	State      string  `json:"state" binding:"required"`
	RecordedAt string  `json:"recorded_at"`
}

type DeleteWeightRecordRequest struct {
	UserID   uint `json:"user_id" binding:"required"`
	RecordID uint `json:"record_id" binding:"required"`
}

func SaveWeightRecord(c *gin.Context) {
	log.Println("[WEIGHT][SAVE][S1] bind request")
	var request SaveWeightRecordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "E_WEIGHT_SAVE_BIND", "参数错误", err)
		return
	}

	if request.UserID == 0 || request.Weight <= 0 {
		respondError(c, http.StatusBadRequest, "E_WEIGHT_SAVE_REQUIRED", "用户和体重必填", nil)
		return
	}

	state := normalizeWeightState(request.State)
	if state == "" {
		respondError(c, http.StatusBadRequest, "E_WEIGHT_SAVE_STATE", "体重状态必须是空腹/正常/饱腹", nil)
		return
	}

	unit := normalizeWeightUnit(request.WeightUnit)
	if unit == "" {
		unit = "kg"
	}
	rawKg := toKg(request.Weight, unit)
	if rawKg <= 0 {
		respondError(c, http.StatusBadRequest, "E_WEIGHT_SAVE_WEIGHT", "体重格式错误", nil)
		return
	}

	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Where("user_id = ?", request.UserID).First(&profile).Error; err != nil {
		respondError(c, http.StatusBadRequest, "E_WEIGHT_SAVE_PROFILE", "请先完成个人资料", err)
		return
	}

	param, err := loadWeightParam(request.UserID, profile.Gender)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "E_WEIGHT_SAVE_PARAM", "读取体重参数失败", err)
		return
	}

	avgFastingKg := computeWeighted7DayKg(request.UserID)
	if avgFastingKg <= 0 {
		avgFastingKg = rawKg
	}

	scale := rawKg / 50.0
	if scale <= 0 {
		scale = 1
	}

	fastingKg := rawKg
	if state == "normal" {
		fastingKg = rawKg - scale*param.NormalDeltaKg
		newNormalDelta := param.NormalDeltaKg*0.9 + (rawKg-avgFastingKg)*0.1
		param.NormalDeltaKg = clampMin(newNormalDelta, 0)
	} else if state == "full" {
		fastingKg = rawKg - scale*(param.NormalDeltaKg+param.FullDeltaKg)
		predictedNormal := avgFastingKg + scale*param.NormalDeltaKg
		newFullDelta := param.FullDeltaKg*0.9 + (rawKg-predictedNormal)*0.1
		param.FullDeltaKg = clampMin(newFullDelta, 0)
	}
	if fastingKg <= 0 {
		fastingKg = rawKg
	}

	recordTime := time.Now()
	if text := strings.TrimSpace(request.RecordedAt); text != "" {
		parsed, parseErr := time.Parse(time.RFC3339, text)
		if parseErr == nil {
			recordTime = parsed
		}
	}
	recordDate := recordTime.Format("2006-01-02")

	record := sqlinit.UserWeightRecord{
		UserID:          request.UserID,
		RecordDate:      recordDate,
		RawWeightKg:     rawKg,
		FastingWeightKg: fastingKg,
		State:           state,
		CreatedAt:       recordTime,
	}
	if err := sqlinit.DB.Create(&record).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "E_WEIGHT_SAVE_RECORD", "保存体重失败", err)
		return
	}

	param.LastComputedAvg = avgFastingKg
	param.LastUpdatedState = state
	param.UpdatedAt = time.Now()
	if err := sqlinit.DB.Save(&param).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "E_WEIGHT_SAVE_PARAM_UPDATE", "更新体重参数失败", err)
		return
	}

	weightedKg := computeWeighted7DayKg(request.UserID)
	if weightedKg <= 0 {
		weightedKg = fastingKg
	}

	updates := map[string]interface{}{
		"weight_kg":     weightedKg,
		"weight_unit":   unit,
		"weight_locked": false,
	}
	if err := sqlinit.DB.Model(&profile).Updates(updates).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "E_WEIGHT_SAVE_PROFILE_UPDATE", "同步体重失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":                true,
		"record_id":         record.ID,
		"display_weight_kg": roundOne(weightedKg),
		"display_weight":    roundOne(fromKg(weightedKg, unit)),
		"weight_unit":       unit,
		"normal_delta_kg":   roundOne(param.NormalDeltaKg),
		"full_delta_kg":     roundOne(param.FullDeltaKg),
	})
}

func GetWeightRecords(c *gin.Context) {
	log.Println("[WEIGHT][RECORDS][S1] parse query")
	userIDText := c.Query("user_id")
	userID64, err := strconv.ParseUint(userIDText, 10, 64)
	if err != nil || userID64 == 0 {
		respondError(c, http.StatusBadRequest, "E_WEIGHT_RECORDS_USER_ID", "user_id参数错误", err)
		return
	}
	userID := uint(userID64)

	limit := 30
	if limitText := strings.TrimSpace(c.Query("limit")); limitText != "" {
		limitParsed, parseErr := strconv.Atoi(limitText)
		if parseErr != nil || limitParsed <= 0 {
			respondError(c, http.StatusBadRequest, "E_WEIGHT_RECORDS_LIMIT", "limit参数错误", parseErr)
			return
		}
		if limitParsed > 200 {
			limitParsed = 200
		}
		limit = limitParsed
	}

	var beforeID uint
	if beforeText := strings.TrimSpace(c.Query("before_id")); beforeText != "" {
		beforeID64, parseErr := strconv.ParseUint(beforeText, 10, 64)
		if parseErr != nil {
			respondError(c, http.StatusBadRequest, "E_WEIGHT_RECORDS_BEFORE_ID", "before_id参数错误", parseErr)
			return
		}
		beforeID = uint(beforeID64)
	}

	unit := "kg"
	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Where("user_id = ?", userID).First(&profile).Error; err == nil {
		normalized := normalizeWeightUnit(profile.WeightUnit)
		if normalized != "" {
			unit = normalized
		}
	}

	query := sqlinit.DB.Where("user_id = ?", userID)
	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}

	var records []sqlinit.UserWeightRecord
	if err := query.Order("id desc").Limit(limit).Find(&records).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "E_WEIGHT_RECORDS_QUERY", "查询体重历史失败", err)
		return
	}

	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		items = append(items, gin.H{
			"id":                record.ID,
			"record_date":       record.RecordDate,
			"recorded_at":       record.CreatedAt,
			"state":             record.State,
			"state_label":       weightStateLabel(record.State),
			"raw_weight_kg":     roundOne(record.RawWeightKg),
			"raw_weight":        roundOne(fromKg(record.RawWeightKg, unit)),
			"fasting_weight_kg": roundOne(record.FastingWeightKg),
			"fasting_weight":    roundOne(fromKg(record.FastingWeightKg, unit)),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"weight_unit": unit,
		"count":       len(items),
		"items":       items,
	})
}

func DeleteWeightRecord(c *gin.Context) {
	log.Println("[WEIGHT][DELETE][S1] parse request")
	var request DeleteWeightRecordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		userIDText := strings.TrimSpace(c.Query("user_id"))
		recordIDText := strings.TrimSpace(c.Query("record_id"))
		if userIDText == "" || recordIDText == "" {
			respondError(c, http.StatusBadRequest, "E_WEIGHT_DELETE_BIND", "参数错误", err)
			return
		}
		userID64, userErr := strconv.ParseUint(userIDText, 10, 64)
		recordID64, recordErr := strconv.ParseUint(recordIDText, 10, 64)
		if userErr != nil || recordErr != nil || userID64 == 0 || recordID64 == 0 {
			respondError(c, http.StatusBadRequest, "E_WEIGHT_DELETE_PARAMS", "user_id或record_id参数错误", err)
			return
		}
		request.UserID = uint(userID64)
		request.RecordID = uint(recordID64)
	}

	if request.UserID == 0 || request.RecordID == 0 {
		respondError(c, http.StatusBadRequest, "E_WEIGHT_DELETE_REQUIRED", "user_id和record_id必填", nil)
		return
	}

	var record sqlinit.UserWeightRecord
	if err := sqlinit.DB.Where("id = ? AND user_id = ?", request.RecordID, request.UserID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusNotFound, "E_WEIGHT_DELETE_NOT_FOUND", "体重记录不存在", err)
			return
		}
		respondError(c, http.StatusInternalServerError, "E_WEIGHT_DELETE_QUERY", "查询体重记录失败", err)
		return
	}

	if err := sqlinit.DB.Delete(&record).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "E_WEIGHT_DELETE_DB", "删除体重记录失败", err)
		return
	}

	displayWeightKg := 0.0
	unit := "kg"
	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Where("user_id = ?", request.UserID).First(&profile).Error; err == nil {
		normalized := normalizeWeightUnit(profile.WeightUnit)
		if normalized != "" {
			unit = normalized
		}
		displayWeightKg = profile.WeightKg

		weightedKg := computeWeighted7DayKg(request.UserID)
		if weightedKg > 0 {
			displayWeightKg = weightedKg
			if err := sqlinit.DB.Model(&profile).Updates(map[string]interface{}{"weight_kg": weightedKg}).Error; err != nil {
				respondError(c, http.StatusInternalServerError, "E_WEIGHT_DELETE_PROFILE_UPDATE", "更新体重失败", err)
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":                true,
		"deleted_id":        request.RecordID,
		"display_weight_kg": roundOne(displayWeightKg),
		"display_weight":    roundOne(fromKg(displayWeightKg, unit)),
		"weight_unit":       unit,
	})
}

func GetWeightDashboard(c *gin.Context) {
	log.Println("[WEIGHT][DASHBOARD][S1] parse query")
	userIDText := c.Query("user_id")
	userID64, err := strconv.ParseUint(userIDText, 10, 64)
	if err != nil || userID64 == 0 {
		respondError(c, http.StatusBadRequest, "E_WEIGHT_DASHBOARD_USER_ID", "user_id参数错误", err)
		return
	}
	userID := uint(userID64)
	viewMode := strings.TrimSpace(c.DefaultQuery("view", "day"))
	if viewMode != "day" && viewMode != "3day" && viewMode != "week" {
		viewMode = "day"
	}

	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		respondError(c, http.StatusBadRequest, "E_WEIGHT_DASHBOARD_PROFILE", "请先完成个人资料", err)
		return
	}

	windowStart := time.Now().AddDate(0, 0, -120).Format("2006-01-02")
	var records []sqlinit.UserWeightRecord
	if err := sqlinit.DB.Where("user_id = ? AND record_date >= ?", userID, windowStart).Order("record_date asc, created_at asc").Find(&records).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "E_WEIGHT_DASHBOARD_RECORDS", "查询体重记录失败", err)
		return
	}

	points, latestFastingKg := buildDashboardPoints(records, viewMode)
	if latestFastingKg <= 0 {
		latestFastingKg = profile.WeightKg
	}
	timelineDates := buildTimelineDates(points, viewMode)
	timelineIndexMap := buildTimelineIndexMap(timelineDates)
	trend := smoothPoints(points)
	yMin, yMax := calcYAxis(points, latestFastingKg)
	unit := normalizeWeightUnit(profile.WeightUnit)
	if unit == "" {
		unit = "kg"
	}
	weighted7DayKg := computeWeightedKgFromLatestDays(records, 7)
	if weighted7DayKg <= 0 {
		weighted7DayKg = profile.WeightKg
	}

	c.JSON(http.StatusOK, gin.H{
		"weight_unit":             unit,
		"display_weight_kg":       roundOne(weighted7DayKg),
		"display_weight":          roundOne(fromKg(weighted7DayKg, unit)),
		"view":                    viewMode,
		"timeline_dates":          timelineDates,
		"timeline_point_count":    len(timelineDates),
		"point_count":             len(points),
		"points":                  convertPointDisplay(points, unit, timelineIndexMap),
		"trend_points":            convertPointDisplay(trend, unit, timelineIndexMap),
		"y_axis_min":              roundOne(fromKg(yMin, unit)),
		"y_axis_max":              roundOne(fromKg(yMax, unit)),
		"reference_latest_weight": roundOne(fromKg(latestFastingKg, unit)),
	})
}

func pickFloat(primary *float64, fallback *float64) float64 {
	if primary != nil {
		return *primary
	}
	if fallback != nil {
		return *fallback
	}
	return 0
}

func normalizeCoreGoals(coreGoal string, coreGoals []string) []string {
	goalSet := map[string]struct{}{}
	for _, raw := range coreGoals {
		goal := strings.TrimSpace(raw)
		if goal != "" {
			goalSet[goal] = struct{}{}
		}
	}
	for _, raw := range strings.Split(coreGoal, ",") {
		goal := strings.TrimSpace(raw)
		if goal != "" {
			goalSet[goal] = struct{}{}
		}
	}
	out := make([]string, 0, len(goalSet))
	for g := range goalSet {
		out = append(out, g)
	}
	return out
}

func viewBucketDays(view string) int {
	switch view {
	case "3day":
		return 3
	case "week":
		return 7
	default:
		return 1
	}
}

func buildTimelineDates(points []weightPoint, view string) []string {
	if len(points) == 0 {
		return []string{}
	}
	startDate, err := time.Parse("2006-01-02", points[0].Date)
	if err != nil {
		return []string{}
	}
	endDate, err := time.Parse("2006-01-02", points[len(points)-1].Date)
	if err != nil {
		endDate = startDate
	}

	stepDays := viewBucketDays(view)
	if stepDays <= 0 {
		stepDays = 1
	}
	minEndDate := startDate.AddDate(0, 0, stepDays*9)
	if endDate.Before(minEndDate) {
		endDate = minEndDate
	}

	result := make([]string, 0, 10)
	for cursor := startDate; !cursor.After(endDate); cursor = cursor.AddDate(0, 0, stepDays) {
		result = append(result, cursor.Format("2006-01-02"))
	}
	if len(result) < 10 {
		cursor := endDate
		for len(result) < 10 {
			cursor = cursor.AddDate(0, 0, stepDays)
			result = append(result, cursor.Format("2006-01-02"))
		}
	}
	return result
}

func buildTimelineIndexMap(dates []string) map[string]int {
	indexMap := make(map[string]int, len(dates))
	for idx, day := range dates {
		indexMap[day] = idx
	}
	return indexMap
}

func containsGoal(goals []string, target string) bool {
	for _, goal := range goals {
		if goal == target {
			return true
		}
	}
	return false
}

func splitCoreGoals(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		goal := strings.TrimSpace(p)
		if goal != "" {
			out = append(out, goal)
		}
	}
	return out
}

func calcAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.Month() < birthDate.Month() || (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}
	if age < 0 {
		return 0
	}
	return age
}

func formatDatePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}

func toAvatarBase64(avatarBytes []byte) string {
	if len(avatarBytes) == 0 {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(avatarBytes)
}

func normalizeWeightUnit(raw string) string {
	unit := strings.ToLower(strings.TrimSpace(raw))
	if unit == "" || unit == "kg" {
		return "kg"
	}
	if unit == "斤" || unit == "jin" {
		return "jin"
	}
	return ""
}

func normalizeWeightState(raw string) string {
	state := strings.ToLower(strings.TrimSpace(raw))
	switch state {
	case "空腹", "fasting":
		return "fasting"
	case "正常", "normal":
		return "normal"
	case "饱腹", "full":
		return "full"
	default:
		return ""
	}
}

func weightStateLabel(state string) string {
	switch normalizeWeightState(state) {
	case "fasting":
		return "空腹"
	case "normal":
		return "正常"
	case "full":
		return "饱腹"
	default:
		return "未知"
	}
}

func toKg(weight float64, unit string) float64 {
	if unit == "jin" {
		return weight / 2.0
	}
	return weight
}

func fromKg(kg float64, unit string) float64 {
	if unit == "jin" {
		return kg * 2.0
	}
	return kg
}

func roundOne(value float64) float64 {
	return math.Round(value*10) / 10
}

func calcBMI(weightKg float64, heightCm float64) float64 {
	if weightKg <= 0 || heightCm <= 0 {
		return 0
	}
	heightM := heightCm / 100.0
	if heightM <= 0 {
		return 0
	}
	return weightKg / (heightM * heightM)
}

func bmiLevel(bmi float64) string {
	if bmi <= 0 {
		return "未知"
	}
	if bmi < 18.5 {
		return "偏瘦"
	}
	if bmi < 24 {
		return "正常"
	}
	if bmi < 28 {
		return "超重"
	}
	return "肥胖"
}

func clampMin(value float64, min float64) float64 {
	if value < min {
		return min
	}
	return value
}

func defaultWeightParams(gender string) (float64, float64) {
	if strings.EqualFold(strings.TrimSpace(gender), "female") || gender == "女" {
		return 0.5, 0.2
	}
	return 0.6, 0.3
}

func ensureWeightParam(userID uint, gender string) error {
	var param sqlinit.UserWeightParam
	err := sqlinit.DB.Where("user_id = ?", userID).First(&param).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	normalDelta, fullDelta := defaultWeightParams(gender)
	param = sqlinit.UserWeightParam{
		UserID:        userID,
		NormalDeltaKg: normalDelta,
		FullDeltaKg:   fullDelta,
		UpdatedAt:     time.Now(),
	}
	return sqlinit.DB.Create(&param).Error
}

func loadWeightParam(userID uint, gender string) (sqlinit.UserWeightParam, error) {
	var param sqlinit.UserWeightParam
	err := sqlinit.DB.Where("user_id = ?", userID).First(&param).Error
	if err == nil {
		if param.NormalDeltaKg <= 0 || param.FullDeltaKg < 0 {
			normalDelta, fullDelta := defaultWeightParams(gender)
			if param.NormalDeltaKg <= 0 {
				param.NormalDeltaKg = normalDelta
			}
			if param.FullDeltaKg < 0 {
				param.FullDeltaKg = fullDelta
			}
		}
		return param, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return param, err
	}
	normalDelta, fullDelta := defaultWeightParams(gender)
	param = sqlinit.UserWeightParam{
		UserID:        userID,
		NormalDeltaKg: normalDelta,
		FullDeltaKg:   fullDelta,
		UpdatedAt:     time.Now(),
	}
	err = sqlinit.DB.Create(&param).Error
	return param, err
}

func computeWeighted7DayKg(userID uint) float64 {
	start := time.Now().AddDate(0, 0, -6).Format("2006-01-02")
	var records []sqlinit.UserWeightRecord
	err := sqlinit.DB.Where("user_id = ? AND record_date >= ?", userID, start).Order("record_date asc, created_at asc").Find(&records).Error
	if err != nil {
		log.Printf("[WEIGHT][CALC][E_WEIGHT_RECORD_QUERY] %v", err)
		return 0
	}
	return computeWeighted7DayFromRecords(records)
}

func resolveProfileWeightedKg(profile *sqlinit.UserProfile) {
	if profile == nil || profile.UserID == 0 {
		return
	}
	weightedKg := computeWeighted7DayKg(profile.UserID)
	if weightedKg <= 0 {
		return
	}
	if math.Abs(profile.WeightKg-weightedKg) < 0.0001 {
		profile.WeightKg = weightedKg
		return
	}
	if err := sqlinit.DB.Model(profile).Updates(map[string]interface{}{"weight_kg": weightedKg}).Error; err != nil {
		log.Printf("[PROFILE][GET][E_PROFILE_SYNC_WEIGHT] user_id=%d err=%v", profile.UserID, err)
		return
	}
	profile.WeightKg = weightedKg
}

func computeWeighted7DayFromRecords(records []sqlinit.UserWeightRecord) float64 {
	daily := buildDailyAverage(records)
	if len(daily) == 0 {
		return 0
	}
	filtered := filterOutlierByMedian(daily)
	if len(filtered) == 0 {
		filtered = daily
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Date < filtered[j].Date })
	weightSum := 0.0
	valueSum := 0.0
	for i, item := range filtered {
		w := math.Log(float64(i) + 2)
		weightSum += w
		valueSum += item.Value * w
	}
	if weightSum == 0 {
		return 0
	}
	return valueSum / weightSum
}

func computeWeightedKgFromLatestDays(records []sqlinit.UserWeightRecord, days int) float64 {
	if days <= 0 {
		return computeWeighted7DayFromRecords(records)
	}
	start := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	filtered := make([]sqlinit.UserWeightRecord, 0, len(records))
	for _, record := range records {
		if record.RecordDate >= start {
			filtered = append(filtered, record)
		}
	}
	return computeWeighted7DayFromRecords(filtered)
}

type weightPoint struct {
	Date  string
	Value float64
}

func buildDailyAverage(records []sqlinit.UserWeightRecord) []weightPoint {
	byDate := map[string][]float64{}
	for _, r := range records {
		if r.FastingWeightKg <= 0 {
			continue
		}
		byDate[r.RecordDate] = append(byDate[r.RecordDate], r.FastingWeightKg)
	}
	points := make([]weightPoint, 0, len(byDate))
	for date, values := range byDate {
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		if len(values) > 0 {
			points = append(points, weightPoint{Date: date, Value: sum / float64(len(values))})
		}
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Date < points[j].Date })
	return points
}

func filterOutlierByMedian(points []weightPoint) []weightPoint {
	if len(points) < 3 {
		return points
	}
	values := make([]float64, 0, len(points))
	for _, p := range points {
		values = append(values, p.Value)
	}
	sort.Float64s(values)
	median := values[len(values)/2]
	if len(values)%2 == 0 {
		median = (values[len(values)/2-1] + values[len(values)/2]) / 2
	}
	out := make([]weightPoint, 0, len(points))
	for _, p := range points {
		if math.Abs(p.Value-median) <= 3 {
			out = append(out, p)
		}
	}
	return out
}

func buildDashboardPoints(records []sqlinit.UserWeightRecord, view string) ([]weightPoint, float64) {
	daily := buildDailyAverage(records)
	if len(daily) == 0 {
		return daily, 0
	}
	latest := daily[len(daily)-1].Value
	if view == "day" {
		return daily, latest
	}

	bucketSize := 3
	if view == "week" {
		bucketSize = 7
	}
	bucketed := make([]weightPoint, 0)
	for i := 0; i < len(daily); i += bucketSize {
		end := i + bucketSize
		if end > len(daily) {
			end = len(daily)
		}
		sum := 0.0
		for j := i; j < end; j++ {
			sum += daily[j].Value
		}
		bucketed = append(bucketed, weightPoint{
			Date:  daily[end-1].Date,
			Value: sum / float64(end-i),
		})
	}
	return bucketed, latest
}

func smoothPoints(points []weightPoint) []weightPoint {
	if len(points) == 0 {
		return points
	}
	window := 3
	out := make([]weightPoint, 0, len(points))
	for i := range points {
		start := i - window + 1
		if start < 0 {
			start = 0
		}
		sum := 0.0
		for j := start; j <= i; j++ {
			sum += points[j].Value
		}
		count := float64(i - start + 1)
		out = append(out, weightPoint{Date: points[i].Date, Value: sum / count})
	}
	return out
}

func calcYAxis(points []weightPoint, latest float64) (float64, float64) {
	if len(points) == 0 {
		if latest <= 0 {
			latest = 60
		}
		return latest - 5, latest + 5
	}
	minVal := points[0].Value
	maxVal := points[0].Value
	for _, p := range points {
		if p.Value < minVal {
			minVal = p.Value
		}
		if p.Value > maxVal {
			maxVal = p.Value
		}
	}
	if latest > 0 {
		if latest < minVal {
			minVal = latest
		}
		if latest > maxVal {
			maxVal = latest
		}
	}
	span := maxVal - minVal
	if span < 10 {
		center := (maxVal + minVal) / 2
		minVal = center - 5
		maxVal = center + 5
		span = 10
	}
	padding := span * 0.1
	return minVal - padding, maxVal + padding
}

func convertPointDisplay(points []weightPoint, unit string, timelineIndexMap map[string]int) []gin.H {
	out := make([]gin.H, 0, len(points))
	for _, p := range points {
		timelineIndex := -1
		if idx, ok := timelineIndexMap[p.Date]; ok {
			timelineIndex = idx
		}
		out = append(out, gin.H{
			"date":           p.Date,
			"weight":         roundOne(fromKg(p.Value, unit)),
			"weight_kg":      roundOne(p.Value),
			"timeline_index": timelineIndex,
		})
	}
	return out
}
