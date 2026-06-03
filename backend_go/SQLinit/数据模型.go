package sqlinit

import "time"

type User struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	Username           string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password           string     `gorm:"size:128;not null" json:"-"`
	AvatarData         []byte     `gorm:"type:longblob" json:"-"`
	MembershipType     string     `gorm:"size:20;default:FREE" json:"membership_type"` // FREE, TRY, PRO_HALF, PRO_ANNUAL
	MembershipExpireAt *time.Time `json:"membership_expire_at"`
	Balance            float64    `gorm:"default:0" json:"balance"`
	FreeTrialUsed      bool       `gorm:"default:false" json:"free_trial_used"`
	CreatedAt          time.Time  `json:"created_at"`
}

type UserProfile struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	UserID                uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	HeightCm              float64    `json:"height_cm"`
	WeightKg              float64    `json:"weight_kg"`
	WeightUnit            string     `gorm:"size:10;default:kg" json:"weight_unit"`
	WeightLocked          bool       `json:"weight_locked"`
	Age                   int        `json:"age"`
	BirthDate             *time.Time `gorm:"type:date" json:"birth_date"`
	Gender                string     `gorm:"size:20" json:"gender"`
	CoreGoal              string     `gorm:"type:text" json:"core_goal"`
	DetailedGoal          string     `gorm:"type:text" json:"detailed_goal"`
	InjuryHistory         string     `gorm:"type:text" json:"injury_history"`
	FavoriteSports        string     `gorm:"type:text" json:"favorite_sports"`
	AddFavoriteToPlan     bool       `json:"add_favorite_to_plan"`
	CurrentExerciseDesc   string     `gorm:"type:text" json:"current_exercise_desc"`
	ExpectedIntensity     string     `gorm:"size:20" json:"expected_intensity"`
	LatestPlanJSON        string     `gorm:"type:longtext" json:"latest_plan_json"`
	LatestPlanGeneratedAt time.Time  `json:"latest_plan_generated_at"`
}

type PlanRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	PlanType  string    `gorm:"size:30;not null" json:"plan_type"`
	PlanJSON  string    `gorm:"type:longtext;not null" json:"plan_json"`
	CreatedAt time.Time `json:"created_at"`
}

type DailyTaskStatus struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	TaskDate  string    `gorm:"size:10;index;not null" json:"task_date"`
	Activity  string    `gorm:"size:100;index;not null" json:"activity"`
	Completed bool      `json:"completed"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessageRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_user_chat_created,priority:1;index;not null" json:"user_id"`
	ChatID    string    `gorm:"size:64;index:idx_user_chat_created,priority:2;not null" json:"chat_id"`
	Role      string    `gorm:"size:20;not null" json:"role"`
	Content   string    `gorm:"type:longtext" json:"content"`
	ImageData []byte    `gorm:"type:longblob" json:"-"`
	JSONData  string    `gorm:"type:longtext" json:"json_data"`
	JSONType  string    `gorm:"size:20" json:"json_type"`
	CreatedAt time.Time `gorm:"index:idx_user_chat_created,priority:3" json:"created_at"`
}

type UserWeightParam struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	NormalDeltaKg    float64   `json:"normal_delta_kg"`
	FullDeltaKg      float64   `json:"full_delta_kg"`
	LastComputedAvg  float64   `json:"last_computed_avg"`
	LastUpdatedState string    `gorm:"size:20" json:"last_updated_state"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserWeightRecord struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index:idx_user_date_created,priority:1;index;not null" json:"user_id"`
	RecordDate      string    `gorm:"size:10;index:idx_user_date_created,priority:2;not null" json:"record_date"`
	RawWeightKg     float64   `json:"raw_weight_kg"`
	FastingWeightKg float64   `json:"fasting_weight_kg"`
	State           string    `gorm:"size:20;index" json:"state"`
	CreatedAt       time.Time `gorm:"index:idx_user_date_created,priority:3" json:"created_at"`
}

type UserMemoryRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_user_memory_updated,priority:1;not null" json:"user_id"`
	Memory    string    `gorm:"type:text;not null" json:"memory"`
	Tag       string    `gorm:"size:30;index" json:"tag"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `gorm:"index:idx_user_memory_updated,priority:2" json:"updated_at"`
}

type UserAttendanceRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index:idx_user_att_date,priority:1;not null" json:"user_id"`
	AttDate   string    `gorm:"size:10;index:idx_user_att_date,priority:2;not null" json:"att_date"` // YYYY-MM-DD
	CreatedAt time.Time `json:"created_at"`
}

type UserAttendanceWeek struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	UserID     uint `gorm:"index:idx_user_week,priority:1;not null" json:"user_id"`
	Year       int  `gorm:"index:idx_user_week,priority:2;not null" json:"year"`
	WeekNumber int  `gorm:"index:idx_user_week,priority:3;not null" json:"week_number"`
	IsFull     bool `json:"is_full"`
	Refunded   bool `gorm:"default:false" json:"refunded"`
}

type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Price     float64   `json:"price"`
	Category  string    `gorm:"size:50;index" json:"category"`
	ImagePath string    `gorm:"size:255" json:"image_path"`
	CreatedAt time.Time `json:"created_at"`
}
