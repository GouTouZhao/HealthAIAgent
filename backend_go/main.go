package main

import (
	"log"

	agentquestion "backend_go/AgentQuestion"
	personinfo "backend_go/PersonInfo"
	sqlinit "backend_go/SQLinit"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("[BOOT][S1] start backend boot")
	log.Println("[BOOT][S2] init db config from env")
	err := sqlinit.InitDB(sqlinit.NewConfigFromEnv())
	if err != nil {
		log.Fatalf("[BOOT][E_BOOT_DB_INIT] 数据库初始化失败: %v", err)
	}
	log.Println("[BOOT][S3] db init success")

	gin.SetMode(gin.DebugMode)
	log.Println("[BOOT][S4] gin mode: debug")
	router := gin.Default()
	log.Println("[BOOT][S5] gin default router created")
	router.Use(cors.Default())
	log.Println("[BOOT][S6] cors middleware attached")

	api := router.Group("/api")
	log.Println("[BOOT][S7] /api group registered")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", personinfo.Register)
			auth.POST("/login", personinfo.Login)
			log.Println("[BOOT][S8] auth routes registered: POST /register, POST /login")
		}

		person := api.Group("/person")
		{
			person.POST("/profile", personinfo.SaveProfile)
			person.GET("/profile", personinfo.GetProfile)
			person.GET("/memory", personinfo.GetMemoryList)
			person.DELETE("/memory", personinfo.DeleteMemory)
			person.POST("/weight-record", personinfo.SaveWeightRecord)
			person.GET("/weight-record", personinfo.GetWeightRecords)
			person.DELETE("/weight-record", personinfo.DeleteWeightRecord)
			person.GET("/weight-dashboard", personinfo.GetWeightDashboard)
			person.GET("/account", personinfo.GetAccount)
			person.POST("/account", personinfo.UpdateAccount)
			person.POST("/reset-password", personinfo.ResetPassword)
			// Membership & Mall
			person.GET("/membership", personinfo.GetMembershipInfo)
			person.POST("/membership/purchase", personinfo.PurchaseMembership)
			person.GET("/mall/products", personinfo.GetProducts)
			person.POST("/balance/withdraw", personinfo.WithdrawBalance)
			log.Println("[BOOT][S9] person routes registered")
		}

		agent := api.Group("/agent")
		{
			agent.POST("/ask", agentquestion.AskAgent)
			agent.POST("/ask-stream", agentquestion.AskAgentStream)
			agent.GET("/chat-history", agentquestion.GetChatHistory)
			agent.DELETE("/chat-history", agentquestion.DeleteChatHistory)
			agent.GET("/plan-history", agentquestion.GetPlanHistory)
			agent.DELETE("/plan-history", agentquestion.DeletePlanHistory)
			agent.POST("/plan-history/replace", agentquestion.ReplacePlanFromHistory)
			agent.GET("/today-plan", agentquestion.GetTodayPlan)
			agent.POST("/toggle-task", agentquestion.ToggleTask)
			log.Println("[BOOT][S10] agent routes registered")
		}
	}

	log.Println("[BOOT][S11] backend_go running at :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("[BOOT][E_BOOT_HTTP_RUN] 服务启动失败: %v", err)
	}
}
