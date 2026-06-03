package personinfo

import (
	sqlinit "backend_go/SQLinit"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RecordAttendance marks today as attended for the user
func RecordAttendance(userID uint) {
	today := time.Now().Format("2006-01-02")
	var existing sqlinit.UserAttendanceRecord
	err := sqlinit.DB.Where("user_id = ? AND att_date = ?", userID, today).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		sqlinit.DB.Create(&sqlinit.UserAttendanceRecord{
			UserID:  userID,
			AttDate: today,
		})
		log.Printf("[Attendance] User %d attended on %s", userID, today)
		// Check if this completes a week
		UpdateWeeklyAttendance(userID)
	}
}

// UpdateWeeklyAttendance checks if the current week's attendance is full
func UpdateWeeklyAttendance(userID uint) {
	now := time.Now()
	// Check the previous week (since the current one isn't finished)
	lastWeek := now.AddDate(0, 0, -7)
	year, week := lastWeek.ISOWeek()

	var existingWeek sqlinit.UserAttendanceWeek
	err := sqlinit.DB.Where("user_id = ? AND year = ? AND week_number = ?", userID, year, week).First(&existingWeek).Error
	if err == nil {
		return // Already processed
	}

	// Get user's current plan
	var profile sqlinit.UserProfile
	if err := sqlinit.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return
	}

	if profile.LatestPlanJSON == "" {
		return
	}

	// Simple check: how many non-rest days?
	// For demo, let's assume 6 days are required if not specified.
	// Actually, point 28 says every day must have activity, but at least one rest day.
	// We'll look for "rest" in the JSON.

	requiredDays := 0
	daysMet := 0

	// Get start of last week (Monday)
	weekday := int(lastWeek.Weekday())
	if weekday == 0 {
		weekday = 7
	} // Sunday
	monday := lastWeek.AddDate(0, 0, -weekday+1)

	for i := 0; i < 7; i++ {
		date := monday.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		dayName := strings.ToLower(date.Weekday().String())

		// Check if it's a rest day in the plan
		isRest := strings.Contains(strings.ToLower(profile.LatestPlanJSON), fmt.Sprintf(`"%s": [{"activity": "rest"`, dayName)) ||
			strings.Contains(strings.ToLower(profile.LatestPlanJSON), fmt.Sprintf(`"%s":[{"activity":"rest"`, dayName))

		if !isRest {
			requiredDays++
			var att sqlinit.UserAttendanceRecord
			if err := sqlinit.DB.Where("user_id = ? AND att_date = ?", userID, dateStr).First(&att).Error; err == nil {
				daysMet++
			}
		}
	}

	isFull := requiredDays > 0 && daysMet >= requiredDays
	newWeek := sqlinit.UserAttendanceWeek{
		UserID:     userID,
		Year:       year,
		WeekNumber: week,
		IsFull:     isFull,
	}
	sqlinit.DB.Create(&newWeek)

	if isFull {
		CheckAndGrantCashback(userID)
	}
}

// CheckAndGrantCashback handles the cashback logic for milestones
func CheckAndGrantCashback(userID uint) {
	var user sqlinit.User
	if err := sqlinit.DB.First(&user, userID).Error; err != nil {
		return
	}

	if user.MembershipType == "FREE" || user.MembershipType == "TRY" {
		return
	}

	var fullWeeksCount int64
	sqlinit.DB.Model(&sqlinit.UserAttendanceWeek{}).Where("user_id = ? AND is_full = ? AND refunded = ?", userID, true, false).Count(&fullWeeksCount)

	var refundPercent float64 = 0
	fee := 0.0

	if user.MembershipType == "PRO_ANNUAL" {
		fee = 88.0
		if fullWeeksCount >= 48 {
			refundPercent = 0.5
		}
		if fullWeeksCount >= 30 && refundPercent == 0 {
			refundPercent = 0.3
		}
		if fullWeeksCount >= 10 && refundPercent == 0 {
			refundPercent = 0.1
		}
	} else if user.MembershipType == "PRO_HALF" {
		fee = 48.0
		if fullWeeksCount >= 24 {
			refundPercent = 0.5
		}
		if fullWeeksCount >= 15 && refundPercent == 0 {
			refundPercent = 0.3
		}
		if fullWeeksCount >= 5 && refundPercent == 0 {
			refundPercent = 0.1
		}
	}

	if refundPercent > 0 {
		refundAmount := fee * refundPercent
		user.Balance += refundAmount
		sqlinit.DB.Save(&user)
		// Mark weeks as refunded to prevent double cashback
		// For simplicity, we'll mark all current full weeks as refunded
		sqlinit.DB.Model(&sqlinit.UserAttendanceWeek{}).Where("user_id = ? AND is_full = ? AND refunded = ?", userID, true, false).Update("refunded", true)
		log.Printf("[Cashback] Granted %.2f to user %d (%s)", refundAmount, userID, user.MembershipType)
	}
}

// GetMembershipInfo returns membership status and attendance progress
func GetMembershipInfo(c *gin.Context) {
	userIDText := c.Query("user_id")
	userID, _ := strconv.ParseUint(userIDText, 10, 64)
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	var user sqlinit.User
	if err := sqlinit.DB.First(&user, uint(userID)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	now := time.Now()
	if user.MembershipType != "FREE" && (user.MembershipExpireAt == nil || !user.MembershipExpireAt.After(now)) {
		user.MembershipType = "FREE"
		user.MembershipExpireAt = nil
		if err := sqlinit.DB.Model(&sqlinit.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"membership_type":      user.MembershipType,
			"membership_expire_at": user.MembershipExpireAt,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to normalize membership"})
			return
		}
	}

	// Record attendance whenever membership info is checked (usually on entry to membership page)
	RecordAttendance(uint(userID))

	// Calculate attendance weeks for cashback
	var fullWeeks []sqlinit.UserAttendanceWeek
	sqlinit.DB.Where("user_id = ? AND is_full = ?", uint(userID), true).Find(&fullWeeks)

	c.JSON(http.StatusOK, gin.H{
		"membership_type":      user.MembershipType,
		"membership_expire_at": user.MembershipExpireAt,
		"balance":              user.Balance,
		"free_trial_used":      user.FreeTrialUsed,
		"full_weeks_count":     len(fullWeeks),
	})
}

// PurchaseMembership handles membership purchase (mock)
func PurchaseMembership(c *gin.Context) {
	var req struct {
		UserID uint   `json:"user_id"`
		Plan   string `json:"plan"` // PRO_HALF, PRO_ANNUAL, TRY
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user sqlinit.User
	if err := sqlinit.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	now := time.Now()
	var expireAt time.Time
	if user.MembershipExpireAt != nil && user.MembershipExpireAt.After(now) {
		expireAt = *user.MembershipExpireAt
	} else {
		expireAt = now
	}

	switch req.Plan {
	case "TRY":
		if user.FreeTrialUsed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Free trial already used"})
			return
		}
		user.MembershipType = "TRY"
		newExpire := expireAt.AddDate(0, 1, 0)
		user.MembershipExpireAt = &newExpire
		user.FreeTrialUsed = true
	case "PRO_HALF":
		user.MembershipType = "PRO_HALF"
		newExpire := expireAt.AddDate(0, 6, 0)
		user.MembershipExpireAt = &newExpire
	case "PRO_ANNUAL":
		user.MembershipType = "PRO_ANNUAL"
		newExpire := expireAt.AddDate(1, 0, 0)
		user.MembershipExpireAt = &newExpire
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan"})
		return
	}

	if err := sqlinit.DB.Model(&sqlinit.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"membership_type":      user.MembershipType,
		"membership_expire_at": user.MembershipExpireAt,
		"free_trial_used":      user.FreeTrialUsed,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update membership"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success", "membership_type": user.MembershipType, "membership_expire_at": user.MembershipExpireAt})
}

// Mall endpoints

// GetProducts returns products by category
func GetProducts(c *gin.Context) {
	category := c.Query("category")
	var products []sqlinit.Product
	query := sqlinit.DB
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	// If no products, return some mock ones as requested (fake demo)
	if len(products) == 0 {
		products = []sqlinit.Product{
			{Name: "筋膜枪", Price: 199, Category: "健身器材", ImagePath: "筋膜枪_199.jpg"},
			{Name: "瑜伽垫", Price: 59, Category: "健身器材", ImagePath: "瑜伽垫_59.jpg"},
			{Name: "乳清蛋白粉", Price: 299, Category: "营养补剂", ImagePath: "乳清蛋白粉_299.jpg"},
			{Name: "运动水壶", Price: 39, Category: "运动装备", ImagePath: "运动水壶_39.jpg"},
		}
	}

	c.JSON(http.StatusOK, products)
}

// WithdrawBalance mock withdrawal
func WithdrawBalance(c *gin.Context) {
	var req struct {
		UserID uint    `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user sqlinit.User
	if err := sqlinit.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	if user.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	user.Balance -= req.Amount
	sqlinit.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Successfully withdrawn %.2f", req.Amount), "balance": user.Balance})
}
