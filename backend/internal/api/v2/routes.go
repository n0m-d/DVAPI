package v2

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/n0m-d/DVAPI/graph"
	"github.com/n0m-d/DVAPI/internal/cache"
	"github.com/n0m-d/DVAPI/internal/handler"
	"github.com/n0m-d/DVAPI/internal/middleware"
	"github.com/n0m-d/DVAPI/internal/service"
	"github.com/vektah/gqlparser/v2/ast"

	ghandler "github.com/99designs/gqlgen/graphql/handler"
)

type Config struct {
	Auth           *handler.AuthHandler
	Users          *handler.UserHandler
	PasswordReset  *handler.PasswordResetHandler
	Admin          *handler.AdminHandler
	JWTSecret      string
	TokenBlacklist cache.TokenBlacklist
	Course         *handler.CourseHandler
	Assignments    *handler.AssignmentHandler
	Learning       *handler.LearningHandler
	Notifications  *handler.NotificationHandler
	Notes          service.NoteService
	RateLimiter    *redis_rate.Limiter
}

func Register(r *gin.RouterGroup, cfg Config) {
	/*
		- Each v2 API route is rate limited to 60 requests per minute.
		- Can be separated into different rate limits for different routes;
		- With different keys for different route groups.
	*/
	r.Use(middleware.RateLimit(cfg.RateLimiter))
	r.POST("/auth/login", cfg.Auth.Login)
	r.POST("/auth/register", cfg.Auth.Register)

	r.POST("/auth/password-reset/request", cfg.PasswordReset.Request)
	r.POST("/auth/password-reset/verify", cfg.PasswordReset.Verify)
	r.POST("/auth/password-reset/confirm", cfg.PasswordReset.Confirm)
	r.GET("/get-uploads", cfg.Assignments.GetFile)

	protected := r.Group("/")
	protected.Use(middleware.JWTVerify(cfg.JWTSecret, cfg.TokenBlacklist))
	{
		protected.POST("/auth/update-password", cfg.Auth.UpdatePassword)
		protected.POST("/auth/logout", cfg.Auth.Logout)
		protected.GET("/users/me", cfg.Users.GetCurrentUser)
		protected.GET("/users/:id", cfg.Users.GetByID)
		protected.PATCH("/users/:id", cfg.Users.UpdateProfile)
		protected.GET("/notifications/stream", cfg.Notifications.Stream)
		protected.GET("/notifications", cfg.Notifications.List)
		protected.POST("/notifications/read-all", cfg.Notifications.MarkAllRead)
		protected.POST("/notifications/:id/read", cfg.Notifications.MarkRead)
		protected.DELETE("/notifications/:id", cfg.Notifications.Delete)
		protected.GET("/courses", cfg.Course.GetCourses)
		protected.GET("/courses/:courseId", cfg.Course.GetByID)

		protected.GET("/assignments/:id", cfg.Assignments.GetByID)
		protected.GET("/stats", cfg.Learning.GetStats)

		studentOnly := protected.Group("/")
		studentOnly.Use(middleware.RoleRestrict("student"))
		{
			studentOnly.GET("/courses/:courseId/assignments", cfg.Assignments.ListByCourse)
			studentOnly.GET("/enrolled-courses", cfg.Course.GetEnrolledCourses)
			studentOnly.GET("/enrolled-courses-count", cfg.Course.GetEnrolledCoursesCount)
			studentOnly.POST("/assignments/:id/submissions", cfg.Assignments.CreateSubmission)
			studentOnly.GET("/assignments/:id/submissions/me", cfg.Assignments.GetMySubmission)
			studentOnly.POST("/courses/:courseId/enrollments/me", cfg.Course.Enroll)
			studentOnly.DELETE("/courses/:courseId/enrollments/me", cfg.Course.Unenroll)
			studentOnly.GET("/courses/:courseId/lessons", cfg.Course.ListStudentLessons)
			studentOnly.PUT("/lessons/:id/progress", cfg.Learning.SetLessonProgress)
			studentOnly.GET("/courses/:courseId/progress", cfg.Learning.GetCourseProgress)
			studentOnly.GET("/courses/:courseId/continue", cfg.Learning.GetNextLesson)
			studentOnly.GET("/grades", cfg.Learning.GetGrades)
			studentOnly.PUT("/submissions/:id", cfg.Learning.Resubmit)
			studentOnly.GET("/submissions/:id/versions", cfg.Learning.ListSubmissionVersions)
			studentOnly.GET("/courses/:courseId/announcements", cfg.Learning.ListStudentAnnouncements)
		}

		instructorOnly := protected.Group("/")
		instructorOnly.Use(middleware.RoleRestrict("instructor"))
		{
			instructorOnly.GET("/my-courses", cfg.Course.ListMyCourses)
			instructorOnly.POST("/courses", cfg.Course.Create)
			instructorOnly.PATCH("/courses/:courseId", cfg.Course.Update)
			instructorOnly.DELETE("/courses/:courseId", cfg.Course.Delete)
			instructorOnly.GET("/my-courses/:courseId/lessons", cfg.Course.ListInstructorLessons)
			instructorOnly.POST("/courses/:courseId/lessons", cfg.Course.CreateLesson)
			instructorOnly.PATCH("/lessons/:id", cfg.Course.UpdateLesson)
			instructorOnly.DELETE("/lessons/:id", cfg.Course.DeleteLesson)
			instructorOnly.GET("/courses/:courseId/analytics", cfg.Learning.GetCourseAnalytics)
			instructorOnly.GET("/my-courses/:courseId/announcements", cfg.Learning.ListInstructorAnnouncements)
			instructorOnly.PATCH("/announcements/:id", cfg.Learning.UpdateAnnouncement)
			instructorOnly.GET("/my-courses/:courseId/assignments", cfg.Assignments.ListCourseAssignments)
			instructorOnly.POST("/assignments", cfg.Assignments.CreateAssignment)
			instructorOnly.POST("/assignment/create", cfg.Assignments.CreateAssignment)
			instructorOnly.PATCH("/assignments/:id", cfg.Assignments.UpdateAssignment)
			instructorOnly.DELETE("/assignments/:id", cfg.Assignments.DeleteAssignment)
			instructorOnly.GET("/courses/:courseId/students", cfg.Course.ListStudents)
			instructorOnly.GET("/submissions/:id", cfg.Assignments.GetSubmissionForInstructor)
			instructorOnly.PATCH("/submissions/:id/grade", cfg.Assignments.GradeSubmission)
		}

		protected.GET("/assignments/:id/submissions", cfg.Assignments.ListSubmissions)
		protected.POST("/courses/:courseId/announcements", cfg.Learning.CreateAnnouncement)
		protected.DELETE("/announcements/:id", cfg.Learning.DeleteAnnouncement)

		adminOnly := protected.Group("/admin")
		adminOnly.Use(middleware.RoleRestrict("admin"))
		{
			adminOnly.GET("/stats", cfg.Users.AdminStats)
			adminOnly.GET("/users", cfg.Users.AdminList)
			adminOnly.POST("/users", cfg.Users.AdminCreate)
			adminOnly.PATCH("/users/:id", cfg.Users.AdminUpdate)
			adminOnly.GET("/logs", cfg.Admin.ViewLogs)
		}
	}

	r.GET("/library", fetchURL)

	registerGraphQL(r, cfg.Notes)
}

func registerGraphQL(r *gin.RouterGroup, notes service.NoteService) {
	srv := ghandler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{NoteService: notes},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{}) //Vuln: Introspection Enabled
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	r.GET("/playground", graphqlPlayground)
	r.Any("/query", graphqlQueryHandler(srv))
}

func graphqlPlayground(c *gin.Context) {
	playground.Handler("GraphQL Playground", "/api/v2/query").ServeHTTP(c.Writer, c.Request)
}

func graphqlQueryHandler(srv *ghandler.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		// gqlgen POST transport only accepts application/json.
		// Many clients (curl, some REST UIs) omit Content-Type.
		if c.Request.Method == http.MethodPost {
			mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
			if err != nil || mediaType == "" || mediaType == "text/plain" {
				c.Request.Header.Set("Content-Type", "application/json")
			}
		}
		srv.ServeHTTP(c.Writer, c.Request)
	}
}

func fetchURL(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty url"})
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var data any
	if json.Unmarshal(body, &data) != nil {
		data = string(body)
	}

	c.JSON(http.StatusOK, gin.H{
		"url":  url,
		"data": data,
	})

}
