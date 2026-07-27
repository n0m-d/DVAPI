package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/n0m-d/DVAPI/internal/config"
	"github.com/n0m-d/DVAPI/internal/database"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/utils"
)

const (
	fakeStudentCount    = 20
	fakeInstructorCount = 10
	courseCount         = 20
)

type seedUser struct {
	name  string
	email string
	role  string
}

func randomPassword() string {
	// Alphanumeric, length 12 — meets auth min length without awkward special chars.
	return gofakeit.Password(true, true, true, false, false, 12)
}

func hashPassword(plain string) (string, error) {
	hash, err := utils.HashPassword(plain)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return hash, nil
}

type seedCourse struct {
	title       string
	slug        string
	description string
	published   bool
	lessons     []string
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) == 0 {
		fmt.Fprintln(os.Stderr, "error: no seed target provided")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  go run ./cmd/seed <target> [target...]")
		fmt.Fprintln(os.Stderr, "  make seed                              # seeds everything")
		fmt.Fprintln(os.Stderr, "  make seed ARGS=\"users courses\"         # seed specific targets")
		fmt.Fprintln(os.Stderr, "  make seed-status                       # show current seed state")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Available targets:")
		fmt.Fprintln(os.Stderr, "  status        show counts and which targets look seeded")
		fmt.Fprintln(os.Stderr, "  users         create admin, instructors, and students")
		fmt.Fprintln(os.Stderr, "  courses       create published/draft courses with lessons")
		fmt.Fprintln(os.Stderr, "  enrollments   enroll students into published courses")
		fmt.Fprintln(os.Stderr, "  assignments   create published assignments on courses")
		fmt.Fprintln(os.Stderr, "  announcements create draft/published announcements on courses")
		fmt.Fprintln(os.Stderr, "  submissions   create (and mostly grade) student submissions")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  make seed ARGS=\"users courses enrollments assignments announcements submissions\"")
		os.Exit(1)
	}

	if err := seed(ctx, db.New(pool), argsWithoutProg); err != nil {
		log.Fatalf("seed failed: %v", err)
	}
}

func seed(ctx context.Context, q *db.Queries, args []string) error {
	for _, arg := range args {
		switch arg {
		case "status":
			if err := printSeedStatus(ctx, q); err != nil {
				return err
			}

		case "users":
			fmt.Println("==== Seeding users ====")
			if done, err := usersAlreadySeeded(ctx, q); err != nil {
				return err
			} else if done {
				fmt.Println("  already seeded — skipping")
				continue
			}
			if err := seedUsers(ctx, q); err != nil {
				return err
			}

		case "courses":
			fmt.Println("==== Seeding courses ====")
			if done, err := coursesAlreadySeeded(ctx, q); err != nil {
				return err
			} else if done {
				fmt.Println("  already seeded — skipping")
				continue
			}
			if err := seedCourses(ctx, q); err != nil {
				return err
			}

		case "enrollments":
			fmt.Println("==== Seeding enrollments ====")
			if done, err := enrollmentsAlreadySeeded(ctx, q); err != nil {
				return err
			} else if done {
				fmt.Println("  already seeded — skipping")
				continue
			}
			if err := seedEnrollments(ctx, q); err != nil {
				return err
			}

		case "assignments":
			fmt.Println("==== Seeding assignments ====")
			if done, err := assignmentsAlreadySeeded(ctx, q); err != nil {
				return err
			} else if done {
				fmt.Println("  already seeded — skipping")
				continue
			}
			if err := seedAssignments(ctx, q); err != nil {
				return err
			}

		case "announcements":
			fmt.Println("==== Seeding announcements ====")
			if done, err := announcementsAlreadySeeded(ctx, q); err != nil {
				return err
			} else if done {
				fmt.Println("  already seeded — skipping")
				continue
			}
			if err := seedAnnouncements(ctx, q); err != nil {
				return err
			}

		case "submissions":
			fmt.Println("==== Seeding assignment submissions ====")
			if done, err := submissionsAlreadySeeded(ctx, q); err != nil {
				return err
			} else if done {
				fmt.Println("  already seeded — skipping")
				continue
			}
			if err := seedSubmissions(ctx, q); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown seed target %q (valid: status, users, courses, enrollments, assignments, announcements, submissions)", arg)
		}
	}

	return nil
}

func printSeedStatus(ctx context.Context, q *db.Queries) error {
	stats, err := q.GetAdminStats(ctx)
	if err != nil {
		return fmt.Errorf("load admin stats: %w", err)
	}

	usersDone, err := usersAlreadySeeded(ctx, q)
	if err != nil {
		return err
	}
	coursesDone, err := coursesAlreadySeeded(ctx, q)
	if err != nil {
		return err
	}
	enrollmentsDone, err := enrollmentsAlreadySeeded(ctx, q)
	if err != nil {
		return err
	}
	assignmentsDone, err := assignmentsAlreadySeeded(ctx, q)
	if err != nil {
		return err
	}
	announcementsDone, err := announcementsAlreadySeeded(ctx, q)
	if err != nil {
		return err
	}
	submissionsDone, err := submissionsAlreadySeeded(ctx, q)
	if err != nil {
		return err
	}

	fmt.Println("==== Seed status ====")
	fmt.Printf("  users         %d  (%s)\n", stats.Users, seededLabel(usersDone))
	fmt.Printf("  students      %d\n", stats.Students)
	fmt.Printf("  instructors   %d\n", stats.Instructors)
	fmt.Printf("  courses       %d  (%s)\n", stats.Courses, seededLabel(coursesDone))
	fmt.Printf("  enrollments   %d  (%s)\n", stats.Enrollments, seededLabel(enrollmentsDone))
	fmt.Printf("  assignments   %d  (%s)\n", stats.Assignments, seededLabel(assignmentsDone))
	fmt.Printf("  announcements %s\n", seededLabel(announcementsDone))
	fmt.Printf("  submissions   %d  (%s)\n", stats.Submissions, seededLabel(submissionsDone))
	return nil
}

func seededLabel(done bool) string {
	if done {
		return "seeded"
	}
	return "not seeded"
}

func usersAlreadySeeded(ctx context.Context, q *db.Queries) (bool, error) {
	for _, email := range []string{
		"admin@schole.com",
		"jane.instructor@schole.com",
		"john.student@schole.com",
	} {
		if _, err := q.GetUserByEmail(ctx, email); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return false, nil
			}
			return false, fmt.Errorf("check user %s: %w", email, err)
		}
	}

	instructors, err := q.ListInstructors(ctx)
	if err != nil {
		return false, fmt.Errorf("list instructors: %w", err)
	}
	students, err := q.ListStudents(ctx)
	if err != nil {
		return false, fmt.Errorf("list students: %w", err)
	}

	return len(instructors) >= 1+fakeInstructorCount && len(students) >= 1+fakeStudentCount, nil
}

func coursesAlreadySeeded(ctx context.Context, q *db.Queries) (bool, error) {
	for _, c := range defaultCourses() {
		if _, err := q.GetCourseBySlug(ctx, c.slug); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return false, nil
			}
			return false, fmt.Errorf("check course %s: %w", c.slug, err)
		}
	}
	return true, nil
}

func enrollmentsAlreadySeeded(ctx context.Context, q *db.Queries) (bool, error) {
	stats, err := q.GetAdminStats(ctx)
	if err != nil {
		return false, fmt.Errorf("load admin stats: %w", err)
	}
	return stats.Enrollments > 0, nil
}

func assignmentsAlreadySeeded(ctx context.Context, q *db.Queries) (bool, error) {
	stats, err := q.GetAdminStats(ctx)
	if err != nil {
		return false, fmt.Errorf("load admin stats: %w", err)
	}
	return stats.Assignments > 0, nil
}

func submissionsAlreadySeeded(ctx context.Context, q *db.Queries) (bool, error) {
	stats, err := q.GetAdminStats(ctx)
	if err != nil {
		return false, fmt.Errorf("load admin stats: %w", err)
	}
	return stats.Submissions > 0, nil
}

func announcementsAlreadySeeded(ctx context.Context, q *db.Queries) (bool, error) {
	courses, err := q.ListPublishedCourses(ctx, db.ListPublishedCoursesParams{
		Limit:  5,
		Offset: 0,
	})
	if err != nil {
		return false, fmt.Errorf("list published courses: %w", err)
	}
	if len(courses) == 0 {
		return false, nil
	}
	for _, course := range courses {
		anns, err := q.ListInstructorAnnouncements(ctx, course.ID)
		if err != nil {
			return false, fmt.Errorf("list announcements for %s: %w", course.Slug, err)
		}
		if len(anns) > 0 {
			return true, nil
		}
	}
	return false, nil
}

func seedUsers(ctx context.Context, q *db.Queries) error {
	created := 0

	fixed := []seedUser{
		{name: "Admin User", email: "admin@schole.com", role: "admin"},
		{name: "Jane Instructor", email: "jane.instructor@schole.com", role: "instructor"},
		{name: "John Doe", email: "john.student@schole.com", role: "student"},
	}

	for _, u := range fixed {
		ok, err := ensureUser(ctx, q, u, randomPassword())
		if err != nil {
			return err
		}
		if ok {
			created++
		}
	}

	n, err := seedPeopleToCount(ctx, q, 1+fakeInstructorCount, "instructor")
	if err != nil {
		return err
	}
	created += n

	n, err = seedPeopleToCount(ctx, q, 1+fakeStudentCount, "student")
	if err != nil {
		return err
	}
	created += n

	fmt.Printf("  users created=%d\n", created)
	return nil
}

func ensureUser(ctx context.Context, q *db.Queries, u seedUser, plainPassword string) (bool, error) {
	if _, err := q.GetUserByEmail(ctx, u.email); err == nil {
		return false, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("check user %s: %w", u.email, err)
	}

	// fmt.Printf("  creating user %s\n", u.email)
	// fmt.Printf("  password %s\n", plainPassword)
	hash, err := hashPassword(plainPassword)
	if err != nil {
		return false, err
	}

	if _, err := createUser(ctx, q, u.email, u.name, u.role, hash); err != nil {
		return false, err
	}
	return true, nil
}

func seedPeopleToCount(ctx context.Context, q *db.Queries, target int, role string) (int, error) {
	var existing int
	switch role {
	case "instructor":
		instructors, err := q.ListInstructors(ctx)
		if err != nil {
			return 0, fmt.Errorf("list instructors: %w", err)
		}
		existing = len(instructors)
	case "student":
		students, err := q.ListStudents(ctx)
		if err != nil {
			return 0, fmt.Errorf("list students: %w", err)
		}
		existing = len(students)
	default:
		return 0, fmt.Errorf("unsupported role for fake users: %s", role)
	}

	needed := target - existing
	if needed <= 0 {
		return 0, nil
	}

	created := 0
	for i := 0; i < needed; i++ {
		p := gofakeit.Person()
		email := strings.ToLower(fmt.Sprintf(
			"%s.%s%d@schole.com",
			p.FirstName,
			p.LastName,
			gofakeit.Number(10, 99),
		))
		fullName := p.FirstName + " " + p.LastName

		if _, err := q.GetUserByEmail(ctx, email); err == nil {
			i--
			continue
		}

		plain := randomPassword()
		hash, err := hashPassword(plain)
		if err != nil {
			return created, err
		}

		if _, err := createUser(ctx, q, email, fullName, role, hash); err != nil {
			return created, fmt.Errorf("fake %s %d: %w", role, i+1, err)
		}
		created++
	}
	return created, nil
}

func createUser(ctx context.Context, q *db.Queries, email, fullName, role, hash string) (db.User, error) {
	return q.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         role,
	})
}

func seedCourses(ctx context.Context, q *db.Queries) error {
	instructors, err := q.ListInstructors(ctx)
	if err != nil {
		return fmt.Errorf("list instructors: %w", err)
	}
	if len(instructors) == 0 {
		return errors.New("no instructors available to assign courses")
	}

	courses := defaultCourses()
	plantCourseIdx, plantInstructor, plantPassword, err := prepareDraftCredentialPlant(ctx, q, instructors, courses)
	if err != nil {
		return err
	}

	created, skipped := 0, 0
	for i, c := range courses {
		if _, err := q.GetCourseBySlug(ctx, c.slug); err == nil {
			skipped++
			continue
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("check course %s: %w", c.slug, err)
		}

		instructor := instructors[i%len(instructors)]
		description := c.description
		if i == plantCourseIdx {
			instructor = plantInstructor
			description = fmt.Sprintf(
				"%s\n\nNOTE TO SELF (remove before publishing): temp account recovery — %s / %s",
				c.description,
				plantInstructor.Email,
				plantPassword,
			)
		}

		course, err := q.CreateCourse(ctx, db.CreateCourseParams{
			InstructorID: instructor.ID,
			Title:        c.title,
			Slug:         c.slug,
			Description:  pgtype.Text{String: description, Valid: true},
			Published:    c.published,
		})
		if err != nil {
			return fmt.Errorf("create course %s: %w", c.slug, err)
		}

		for j, lessonTitle := range c.lessons {
			content := gofakeit.Paragraph(2, 6, 15, "\n\n") + gofakeit.LoremIpsumParagraph(1, 3, 12, " ")

			if _, err := q.CreateLesson(ctx, db.CreateLessonParams{
				CourseID:  course.ID,
				Title:     lessonTitle,
				SortOrder: int32(j + 1),
				Content:   pgtype.Text{String: content, Valid: true},
			}); err != nil {
				return fmt.Errorf("create lesson for %s: %w", c.slug, err)
			}
		}
		created++
	}

	fmt.Printf("  courses created=%d skipped=%d\n", created, skipped)
	return nil
}

// prepareDraftCredentialPlant picks a random draft course + instructor and resets that
// instructor's password so it matches the note planted in the course description.
func prepareDraftCredentialPlant(
	ctx context.Context,
	q *db.Queries,
	instructors []db.User,
	courses []seedCourse,
) (courseIdx int, instructor db.User, password string, err error) {
	draftIdxs := make([]int, 0, len(courses))
	for i, c := range courses {
		if !c.published {
			draftIdxs = append(draftIdxs, i)
		}
	}
	if len(draftIdxs) == 0 {
		return -1, db.User{}, "", errors.New("no draft courses available for credential plant")
	}

	courseIdx = draftIdxs[gofakeit.Number(0, len(draftIdxs)-1)]
	instructor = instructors[gofakeit.Number(0, len(instructors)-1)]
	password = randomPassword()

	hash, err := hashPassword(password)
	if err != nil {
		return -1, db.User{}, "", err
	}
	if err := q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           instructor.ID,
		PasswordHash: hash,
	}); err != nil {
		return -1, db.User{}, "", fmt.Errorf("update instructor password: %w", err)
	}

	return courseIdx, instructor, password, nil
}

func defaultCourses() []seedCourse {
	courses := []seedCourse{
		{
			title:       "Go Fundamentals",
			slug:        "go-fundamentals",
			description: "Learn Go syntax, types, and idiomatic patterns from scratch.",
			published:   true,
			lessons:     []string{"Hello Go", "Structs & Interfaces", "Concurrency Basics"},
		},
		{
			title:       "Building APIs with Gin",
			slug:        "building-apis-with-gin",
			description: "Design production HTTP APIs using Gin, middleware, and validation.",
			published:   true,
			lessons:     []string{"Routing", "Middleware", "Error Handling"},
		},
		{
			slug:        "postgresql-for-backend-devs",
			title:       "PostgreSQL for Backend Devs",
			description: "Indexes, transactions, and query design for application developers.",
			published:   true,
			lessons:     []string{"Schema Design", "Indexes", "Transactions"},
		},
		{
			title:       "SQLC Type-Safe Queries",
			slug:        "sqlc-type-safe-queries",
			description: "Write SQL, generate Go, and keep your DB layer honest.",
			published:   true,
			lessons:     []string{"Queries", "Codegen", "Repositories"},
		},
		{
			title:       "Redis Caching Patterns",
			slug:        "redis-caching-patterns",
			description: "Cache-aside, TTLs, and invalidation strategies for APIs.",
			published:   true,
			lessons:     []string{"Cache-Aside", "TTL Design", "Invalidation"},
		},
		{
			title:       "JWT Auth Deep Dive",
			slug:        "jwt-auth-deep-dive",
			description: "Access tokens, logout blacklists, and secure auth flows.",
			published:   true,
			lessons:     []string{"Claims", "Verification", "Blacklisting"},
		},
		{
			title:       "Testing Go HTTP Handlers",
			slug:        "testing-go-http-handlers",
			description: "Table-driven tests for handlers and services without a database.",
			published:   true,
			lessons:     []string{"httptest", "Mocks", "Coverage"},
		},
		{
			title:       "Docker Compose for Local Dev",
			slug:        "docker-compose-local-dev",
			description: "Spin up Postgres, Redis, and friends for day-to-day development.",
			published:   true,
			lessons:     []string{"Compose Basics", "Volumes", "Healthchecks"},
		},
		{
			title:       "Advanced Go Concurrency",
			slug:        "advanced-go-concurrency",
			description: "Worker pools, contexts, and avoiding data races.",
			published:   false,
			lessons:     []string{"Contexts", "Channels", "Sync Primitives"},
		},
		{
			title:       "Observability with slog",
			slug:        "observability-with-slog",
			description: "Structured logs, request IDs, and useful production signal.",
			published:   false,
			lessons:     []string{"Handlers", "Context Fields", "Levels"},
		},
		{
			title:       "Observability with slog",
			slug:        "observability-with-slog",
			description: "Structured logs, request IDs, and useful production signal.",
			published:   false,
			lessons:     []string{"Handlers", "Context Fields", "Levels"},
		},
	}

	for i := range courses {
		// Leave draft descriptions short/clean — one of them gets the credential note.
		if !courses[i].published {
			continue
		}
		courses[i].description += " " + gofakeit.LoremIpsumParagraph(1, 3, 12, " ")
	}

	return courses
}

func seedEnrollments(ctx context.Context, q *db.Queries) error {
	students, err := q.ListStudents(ctx)
	if err != nil {
		return fmt.Errorf("list students: %w", err)
	}
	if len(students) == 0 {
		return errors.New("no students available to enroll")
	}

	listed, err := q.ListPublishedCourses(ctx, db.ListPublishedCoursesParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("list published courses: %w", err)
	}

	courses := make([]db.Course, 0, len(listed))
	for _, course := range listed {
		if course.Published {
			courses = append(courses, course)
		}
	}
	if len(courses) == 0 {
		return errors.New("no published courses available to enroll into")
	}

	created := 0
	skipped := 0

	// Fixed student gets a predictable set of published courses.
	if john, err := q.GetUserByEmail(ctx, "john.student@schole.com"); err == nil {
		limit := 3
		if limit > len(courses) {
			limit = len(courses)
		}
		for _, course := range courses[:limit] {
			ok, err := ensureEnrollment(ctx, q, john, course)
			if err != nil {
				return err
			}
			if ok {
				created++
			} else {
				skipped++
			}
		}
	}

	for _, student := range students {
		if student.Email == "john.student@schole.com" {
			continue
		}

		n := gofakeit.Number(1, min(3, len(courses)))
		picked := pickCourseIndexes(len(courses), n)
		for _, idx := range picked {
			ok, err := ensureEnrollment(ctx, q, student, courses[idx])
			if err != nil {
				return err
			}
			if ok {
				created++
			} else {
				skipped++
			}
		}
	}

	fmt.Printf("  enrollments created=%d skipped=%d (published courses=%d)\n", created, skipped, len(courses))
	return nil
}

func ensureEnrollment(ctx context.Context, q *db.Queries, student db.User, course db.Course) (bool, error) {
	if !course.Published {
		return false, nil
	}

	_, err := q.GetEnrollment(ctx, db.GetEnrollmentParams{
		UserID:   student.ID,
		CourseID: course.ID,
	})
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("check enrollment %s/%s: %w", student.Email, course.Slug, err)
	}

	if _, err := q.CreateEnrollment(ctx, db.CreateEnrollmentParams{
		UserID:   student.ID,
		CourseID: course.ID,
	}); err != nil {
		return false, fmt.Errorf("create enrollment %s -> %s: %w", student.Email, course.Slug, err)
	}

	return true, nil
}

func pickCourseIndexes(total, n int) []int {
	if n > total {
		n = total
	}
	indexes := make([]int, total)
	for i := range indexes {
		indexes[i] = i
	}
	gofakeit.ShuffleInts(indexes)
	return indexes[:n]
}

func seedAssignments(ctx context.Context, q *db.Queries) error {
	jane, err := q.GetUserByEmail(ctx, "jane.instructor@schole.com")
	if err != nil {
		return fmt.Errorf("load instructor jane: %w", err)
	}

	courses, err := q.ListPublishedCourses(ctx, db.ListPublishedCoursesParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("list published courses: %w", err)
	}
	if len(courses) == 0 {
		return errors.New("no published courses available for assignments")
	}

	templates := []struct {
		title       string
		description string
		daysUntil   int
	}{
		{"Week 1 Checkpoint", "Submit a short write-up covering this week's concepts.", 14},
		{"Practical Exercise", "Upload your completed exercise file and notes.", 21},
		{"Final Project Draft", "Share a draft of your final project for feedback.", 30},
	}

	created, skipped := 0, 0
	limit := min(5, len(courses))
	for i := 0; i < limit; i++ {
		course := courses[i]
		tpl := templates[i%len(templates)]
		title := fmt.Sprintf("%s — %s", course.Title, tpl.title)

		_, err := q.GetAssignmentByCourseAndTitle(ctx, db.GetAssignmentByCourseAndTitleParams{
			CourseID: course.ID,
			Title:    title,
		})
		if err == nil {
			skipped++
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("check assignment %s: %w", title, err)
		}

		createdBy := jane.ID
		if course.InstructorID.Valid {
			createdBy = course.InstructorID
		}

		if _, err := q.CreateAssignment(ctx, db.CreateAssignmentParams{
			CourseID:    course.ID,
			Title:       title,
			Description: pgtype.Text{String: tpl.description, Valid: true},
			DueDate:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, tpl.daysUntil), Valid: true},
			Status:      "published",
			CreatedBy:   createdBy,
		}); err != nil {
			return fmt.Errorf("create assignment %s: %w", title, err)
		}
		created++
	}

	fmt.Printf("  assignments created=%d skipped=%d\n", created, skipped)
	return nil
}

func seedAnnouncements(ctx context.Context, q *db.Queries) error {
	jane, err := q.GetUserByEmail(ctx, "jane.instructor@schole.com")
	if err != nil {
		return fmt.Errorf("load instructor jane: %w", err)
	}

	courses, err := q.ListPublishedCourses(ctx, db.ListPublishedCoursesParams{
		Limit:  5,
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("list published courses: %w", err)
	}
	if len(courses) == 0 {
		return errors.New("no published courses available for announcements")
	}

	templates := []struct {
		title   string
		content string
		status  db.AnnouncementStatus
	}{
		{
			title:   "Welcome to the course",
			content: "Glad you're here. Review the syllabus and introduce yourself in the discussion space.",
			status:  db.AnnouncementStatusPublished,
		},
		{
			title:   "Week 1 checklist",
			content: "Complete the first lesson, skim the reading list, and note any questions for office hours.",
			status:  db.AnnouncementStatusPublished,
		},
		{
			title:   "Office hours this week",
			content: "Drop-in help is available mid-week. Bring specific blockers if you can.",
			status:  db.AnnouncementStatusPublished,
		},
		{
			title:   "Syllabus tweak (draft)",
			content: "Considering a due-date shift for the mid-unit checkpoint. Not final — do not rely on this yet.",
			status:  db.AnnouncementStatusDraft,
		},
		{
			title:   "Exam logistics (draft)",
			content: "Placeholder for exam format, allowed materials, and grading weights. Publish once confirmed.",
			status:  db.AnnouncementStatusDraft,
		},
	}

	created, skipped := 0, 0
	for _, course := range courses {
		existing, err := q.ListInstructorAnnouncements(ctx, course.ID)
		if err != nil {
			return fmt.Errorf("list announcements for %s: %w", course.Slug, err)
		}
		existingTitles := make(map[string]struct{}, len(existing))
		for _, a := range existing {
			existingTitles[a.Title] = struct{}{}
		}

		createdBy := jane.ID
		if course.InstructorID.Valid {
			createdBy = course.InstructorID
		}

		for _, tpl := range templates {
			if _, ok := existingTitles[tpl.title]; ok {
				skipped++
				continue
			}
			if _, err := q.CreateAnnouncement(ctx, db.CreateAnnouncementParams{
				CourseID:  course.ID,
				Title:     tpl.title,
				Content:   tpl.content,
				Status:    tpl.status,
				CreatedBy: createdBy,
			}); err != nil {
				return fmt.Errorf("create announcement %s/%s: %w", course.Slug, tpl.title, err)
			}
			created++
		}
	}

	fmt.Printf("  announcements created=%d skipped=%d\n", created, skipped)
	return nil
}

func seedSubmissions(ctx context.Context, q *db.Queries) error {
	courses, err := q.ListPublishedCourses(ctx, db.ListPublishedCoursesParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("list published courses: %w", err)
	}
	if len(courses) == 0 {
		return errors.New("no published courses available for submissions")
	}

	created, skipped, graded := 0, 0, 0

	for _, course := range courses {
		assignments, err := q.ListPublishedAssignmentsByCourse(ctx, course.ID)
		if err != nil {
			return fmt.Errorf("list assignments for %s: %w", course.Slug, err)
		}
		if len(assignments) == 0 {
			continue
		}

		students, err := q.ListEnrolledStudentsByCourse(ctx, db.ListEnrolledStudentsByCourseParams{
			CourseID:    course.ID,
			Name:        "",
			OffsetCount: 0,
			LimitCount:  100,
		})
		if err != nil {
			return fmt.Errorf("list enrolled students for %s: %w", course.Slug, err)
		}
		if len(students) == 0 {
			continue
		}

		for _, assignment := range assignments {
			for _, student := range students {
				ok, wasGraded, err := ensureSubmission(ctx, q, course.ID, assignment, student)
				if err != nil {
					return err
				}
				if ok {
					created++
					if wasGraded {
						graded++
					}
				} else {
					skipped++
				}
			}
		}
	}

	if created == 0 && skipped == 0 {
		return errors.New("no enrolled students on courses with published assignments")
	}

	fmt.Printf("  submissions created=%d skipped=%d graded=%d\n", created, skipped, graded)
	return nil
}

func ensureSubmission(
	ctx context.Context,
	q *db.Queries,
	courseID pgtype.UUID,
	assignment db.Assignment,
	student db.ListEnrolledStudentsByCourseRow,
) (created bool, graded bool, err error) {
	_, err = q.GetSubmissionByAssignmentAndStudent(ctx, db.GetSubmissionByAssignmentAndStudentParams{
		AssignmentID: assignment.ID,
		StudentID:    student.ID,
	})
	if err == nil {
		return false, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, false, fmt.Errorf("check submission %s/%s: %w", student.Email, assignment.Title, err)
	}

	fileName := gofakeit.UUID() + ".md"
	filePath := fmt.Sprintf("uploads/%s/%s/%s", courseID.String(), assignment.ID.String(), fileName)

	submission, err := q.CreateAssignmentSubmission(ctx, db.CreateAssignmentSubmissionParams{
		AssignmentID:   assignment.ID,
		StudentID:      student.ID,
		SubmissionText: pgtype.Text{String: gofakeit.Paragraph(2, 5, 12, " "), Valid: true},
		FilePath:       pgtype.Text{String: filePath, Valid: true},
		FileName:       pgtype.Text{String: fileName, Valid: true},
	})
	if err != nil {
		return false, false, fmt.Errorf("create submission %s -> %s: %w", student.Email, assignment.Title, err)
	}

	// Grade most submissions so dashboards and instructor views have data.
	if gofakeit.Number(1, 100) <= 70 {
		grade := int32(gofakeit.Number(55, 100))
		feedback := gofakeit.Sentence(8)
		if _, err := q.GradeSubmission(ctx, db.GradeSubmissionParams{
			Grade:    pgtype.Int4{Int32: grade, Valid: true},
			Feedback: pgtype.Text{String: feedback, Valid: true},
			ID:       submission.ID,
		}); err != nil {
			return true, false, fmt.Errorf("grade submission %s: %w", submission.ID.String(), err)
		}
		return true, true, nil
	}

	return true, false, nil
}
