package agentquestion

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	sqlinit "backend_go/SQLinit"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type activityItem struct {
	Activity          string `json:"activity"`
	Sets              *int   `json:"sets,omitempty"`
	Reps              *int   `json:"reps,omitempty"`
	DurationMin       *int   `json:"duration_min,omitempty"`
	Intensity         string `json:"intensity,omitempty"`
	SetsReps          string `json:"sets_reps,omitempty"`
	DayIntensity      string `json:"day_intensity,omitempty"`
	EstimatedDuration string `json:"estimated_duration,omitempty"`
	RestInterval      string `json:"rest_interval,omitempty"`
}

type dayPlan struct {
	Activities        []activityItem `json:"activities"`
	DayIntensity      string         `json:"day_intensity"`
	EstimatedDuration string         `json:"estimated_duration"`
	RestInterval      string         `json:"rest_interval"`
}

type monthPlan struct {
	Month      int                       `json:"month"`
	WeeklyPlan map[string][]activityItem `json:"weekly_plan"`
}

type planShape struct {
	Months []monthPlan `json:"months"`
}

type monthPlanV2 struct {
	Month      int                `json:"month"`
	WeeklyPlan map[string]dayPlan `json:"weekly_plan"`
}

type planShapeV2 struct {
	Months []monthPlanV2 `json:"months"`
}

func normalizeTodayTasksFromDayPlan(plan dayPlan) []activityItem {
	tasks := make([]activityItem, 0, len(plan.Activities))
	for _, item := range plan.Activities {
		task := item
		task.DayIntensity = plan.DayIntensity
		task.EstimatedDuration = plan.EstimatedDuration
		task.RestInterval = plan.RestInterval
		tasks = append(tasks, task)
	}
	return tasks
}

type toggleTaskRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`
	TaskDate  string `json:"task_date" binding:"required"`
	Activity  string `json:"activity" binding:"required"`
	Completed bool   `json:"completed"`
}

func GetTodayPlan(c *gin.Context) {
	log.Println("[TODAY_PLAN][S1] parse query user_id")
	userIDText := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDText, 10, 64)
	if err != nil || userID == 0 {
		respondAgentError(c, http.StatusBadRequest, "E_TODAY_PLAN_USER_ID", "user_id参数错误", err)
		return
	}
	log.Printf("[TODAY_PLAN][S2] load profile user_id=%d", userID)

	taskDate := time.Now().Format("2006-01-02")
	emptyResp := gin.H{
		"task_date": taskDate,
		"tasks":     []activityItem{},
		"completed": 0,
		"total":     0,
	}

	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[TODAY_PLAN][S3] profile not found user_id=%d", userID)
			c.JSON(http.StatusOK, emptyResp)
			return
		}
		respondAgentError(c, http.StatusInternalServerError, "E_TODAY_PLAN_PROFILE_QUERY", "查询计划失败", err)
		return
	}
	if profile.LatestPlanJSON == "" {
		log.Printf("[TODAY_PLAN][S3] latest plan empty user_id=%d", userID)
		c.JSON(http.StatusOK, emptyResp)
		return
	}

	var planV2 planShapeV2
	if err := json.Unmarshal([]byte(profile.LatestPlanJSON), &planV2); err != nil {
		log.Printf("[TODAY_PLAN][S4] latest plan json invalid user_id=%d err=%v", userID, err)
		c.JSON(http.StatusOK, emptyResp)
		return
	}

	weekdayMap := map[time.Weekday]string{
		time.Monday:    "monday",
		time.Tuesday:   "tuesday",
		time.Wednesday: "wednesday",
		time.Thursday:  "thursday",
		time.Friday:    "friday",
		time.Saturday:  "saturday",
		time.Sunday:    "sunday",
	}
	dayKey := weekdayMap[time.Now().Weekday()]

	monthIndex := 0
	if len(planV2.Months) > 0 {
		elapsedMonths := int(time.Since(profile.LatestPlanGeneratedAt).Hours() / 24 / 30)
		monthIndex = elapsedMonths
		if monthIndex < 0 {
			monthIndex = 0
		}
		if monthIndex >= len(planV2.Months) {
			monthIndex = len(planV2.Months) - 1
		}
	}

	todayTasks := []activityItem{}
	if len(planV2.Months) > 0 {
		todayDayPlan := planV2.Months[monthIndex].WeeklyPlan[dayKey]
		todayTasks = normalizeTodayTasksFromDayPlan(todayDayPlan)
	} else {
		var legacyPlan planShape
		if err := json.Unmarshal([]byte(profile.LatestPlanJSON), &legacyPlan); err == nil && len(legacyPlan.Months) > 0 {
			if monthIndex >= len(legacyPlan.Months) {
				monthIndex = len(legacyPlan.Months) - 1
			}
			todayTasks = legacyPlan.Months[monthIndex].WeeklyPlan[dayKey]
		}
	}

	statuses := []sqlinit.DailyTaskStatus{}
	if err := sqlinit.DB.Where("user_id = ? AND task_date = ?", userID, taskDate).Find(&statuses).Error; err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_TODAY_PLAN_STATUS_QUERY", "查询任务状态失败", err)
		return
	}
	statusMap := map[string]bool{}
	for _, item := range statuses {
		statusMap[item.Activity] = item.Completed
	}

	completedCount := 0
	for _, item := range todayTasks {
		if statusMap[item.Activity] {
			completedCount++
		}
	}
	log.Printf("[TODAY_PLAN][S5] response user_id=%d task_date=%s total=%d completed=%d", userID, taskDate, len(todayTasks), completedCount)

	c.JSON(http.StatusOK, gin.H{
		"task_date": taskDate,
		"tasks":     todayTasks,
		"completed": completedCount,
		"total":     len(todayTasks),
	})
}

func ToggleTask(c *gin.Context) {
	log.Println("[TOGGLE_TASK][S1] bind request")
	var request toggleTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondAgentError(c, http.StatusBadRequest, "E_TOGGLE_TASK_BIND", "参数错误", err)
		return
	}

	request.Activity = strings.TrimSpace(request.Activity)
	if request.Activity == "" {
		respondAgentError(c, http.StatusBadRequest, "E_TOGGLE_TASK_ACTIVITY", "activity参数错误", nil)
		return
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(request.TaskDate)); err != nil {
		respondAgentError(c, http.StatusBadRequest, "E_TOGGLE_TASK_DATE", "task_date格式错误，应为YYYY-MM-DD", err)
		return
	}
	log.Printf("[TOGGLE_TASK][S2] upsert status user_id=%d date=%s activity=%s completed=%t", request.UserID, request.TaskDate, request.Activity, request.Completed)

	status := sqlinit.DailyTaskStatus{
		UserID:    request.UserID,
		TaskDate:  request.TaskDate,
		Activity:  request.Activity,
		Completed: request.Completed,
	}
	if err := sqlinit.DB.Where("user_id = ? AND task_date = ? AND activity = ?", request.UserID, request.TaskDate, request.Activity).Assign(status).FirstOrCreate(&status).Error; err != nil {
		respondAgentError(c, http.StatusInternalServerError, "E_TOGGLE_TASK_UPSERT", "更新失败", err)
		return
	}
	log.Printf("[TOGGLE_TASK][S3] upsert success user_id=%d", request.UserID)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
