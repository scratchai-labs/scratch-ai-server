package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/assignment"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/auth"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/dashboard"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/hint"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/progress"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/sb3"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/student"
)

const teacherContextKey = "teacher"
const studentContextKey = "student"

func NewRouter(cfg config.Config) (http.Handler, error) {
	store, err := memory.NewStore(cfg)
	if err != nil {
		return nil, err
	}
	authService := auth.NewService(store)
	studentService := student.NewService(store)
	assignmentService := assignment.NewService(store, sb3.NewAnalyzer(), sb3.NewLocalStorage(cfg.SB3StorageDir))
	progressService := progress.NewService(store)
	hintService := hint.NewService(store, hint.NewDeepSeekProvider(cfg.DeepSeek))
	dashboardService := dashboard.NewService(store)

	authRoutes := &authHandler{service: authService}
	studentRoutes := &studentHandler{
		studentService:    studentService,
		assignmentService: assignmentService,
		progressService:   progressService,
		hintService:       hintService,
	}
	assignmentRoutes := &assignmentHandler{
		assignmentService: assignmentService,
	}
	dashboardRoutes := &dashboardHandler{
		dashboardService: dashboardService,
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(allowCORS())

	engine.GET("/health", handleHealth)
	engine.POST("/api/teacher/register", authRoutes.handleTeacherRegister)
	engine.POST("/api/teacher/login", authRoutes.handleTeacherLogin)
	engine.POST("/api/student/login", studentRoutes.handleStudentLogin)

	teacherGroup := engine.Group("/api/teacher")
	teacherGroup.Use(requireTeacher(authService))
	teacherGroup.GET("/me", authRoutes.handleTeacherMe)
	teacherGroup.POST("/logout", authRoutes.handleTeacherLogout)
	teacherGroup.GET("/students", studentRoutes.handleTeacherStudents)
	teacherGroup.POST("/students", studentRoutes.handleTeacherStudents)
	teacherGroup.POST("/students/batch", studentRoutes.handleTeacherStudentsBatch)
	teacherGroup.POST("/students/:id/reset-password", studentRoutes.handleTeacherStudentPasswordReset)
	teacherGroup.GET("/assignments", assignmentRoutes.handleTeacherAssignmentsList)
	teacherGroup.POST("/assignments", assignmentRoutes.handleTeacherAssignments)
	teacherGroup.GET("/assignments/:id", assignmentRoutes.handleTeacherAssignmentDetail)
	teacherGroup.GET("/assignments/:id/analysis", assignmentRoutes.handleTeacherAssignmentAnalysis)
	teacherGroup.POST("/assignments/:id/assign-students", assignmentRoutes.handleTeacherAssignmentStudents)
	teacherGroup.POST("/assignments/:id/publish", assignmentRoutes.handleTeacherAssignmentPublish)
	teacherGroup.POST("/assignments/:id/archive", assignmentRoutes.handleTeacherAssignmentArchive)
	teacherGroup.GET("/dashboard/assignments/:id/live", dashboardRoutes.handleTeacherLiveDashboard)
	teacherGroup.GET("/dashboard/students/:id/history", dashboardRoutes.handleTeacherStudentHistory)

	studentGroup := engine.Group("/api/student")
	studentGroup.Use(requireStudent(studentService))
	studentGroup.GET("/me", studentRoutes.handleStudentMe)
	studentGroup.POST("/logout", studentRoutes.handleStudentLogout)
	studentGroup.GET("/assignments", studentRoutes.handleStudentAssignments)
	studentGroup.GET("/assignments/:id", studentRoutes.handleStudentAssignmentDetail)
	studentGroup.POST("/assignments/:id/progress", studentRoutes.handleStudentAssignmentProgress)
	studentGroup.POST("/assignments/:id/hints", studentRoutes.handleStudentAssignmentHint)

	return engine, nil
}

func requireTeacher(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		teacherRecord, err := authService.TeacherFromBearer(c.GetHeader("Authorization"))
		if err != nil {
			status := 500
			if err == auth.ErrUnauthorized {
				status = 401
			}
			writeJSONError(c, status, err.Error())
			c.Abort()
			return
		}

		c.Set(teacherContextKey, teacherRecord)
		c.Next()
	}
}

func requireStudent(studentService *student.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		studentRecord, err := studentService.StudentFromBearer(c.GetHeader("Authorization"))
		if err != nil {
			status := 500
			if err == student.ErrUnauthorized {
				status = 401
			}
			writeJSONError(c, status, err.Error())
			c.Abort()
			return
		}

		c.Set(studentContextKey, studentRecord)
		c.Next()
	}
}

func mustTeacher(c *gin.Context) memory.Teacher {
	value, ok := c.Get(teacherContextKey)
	if !ok {
		writeJSONError(c, 401, "missing or invalid bearer token")
		c.Abort()
		return memory.Teacher{}
	}

	teacherRecord, ok := value.(memory.Teacher)
	if !ok {
		writeJSONError(c, 500, "teacher context decode failed")
		c.Abort()
		return memory.Teacher{}
	}
	return teacherRecord
}

func mustStudent(c *gin.Context) memory.Student {
	value, ok := c.Get(studentContextKey)
	if !ok {
		writeJSONError(c, 401, "missing or invalid bearer token")
		c.Abort()
		return memory.Student{}
	}

	studentRecord, ok := value.(memory.Student)
	if !ok {
		writeJSONError(c, 500, "student context decode failed")
		c.Abort()
		return memory.Student{}
	}
	return studentRecord
}
