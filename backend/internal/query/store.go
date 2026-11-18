package query

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/n0m-d/DVAPI/internal/db"
	"github.com/n0m-d/DVAPI/internal/domain"
)

type Store interface {
	AdminStats(ctx context.Context) (domain.AdminStats, error)
	ListUsers(ctx context.Context, search, role string, limit, offset int) ([]domain.User, error)
	CountUsers(ctx context.Context, search, role string) (int, error)

	InstructorStats(ctx context.Context, instructorID uuid.UUID) (domain.InstructorStats, error)
	StudentStats(ctx context.Context, studentID uuid.UUID) (domain.StudentStats, error)
	StudentGrades(ctx context.Context, studentID uuid.UUID) (domain.StudentGradesDashboard, error)
	CourseAnalytics(ctx context.Context, courseID uuid.UUID) (domain.CourseAnalytics, error)
	CourseProgress(ctx context.Context, studentID, courseID uuid.UUID) (domain.CourseProgress, error)

	Course(ctx context.Context, id uuid.UUID) (domain.Course, error)
	IsEnrolled(ctx context.Context, studentID, courseID uuid.UUID) (bool, error)

	Courses(ctx context.Context, published bool, title string, limit, offset int) ([]domain.Course, error)
	CountCourses(ctx context.Context, published bool, title string) (int, error)
	EnrolledCourses(ctx context.Context, studentID uuid.UUID, title string, limit, offset int) ([]domain.Course, error)
	CountEnrolledCourses(ctx context.Context, studentID uuid.UUID, title string) (int, error)
	InstructorCourses(ctx context.Context, instructorID uuid.UUID, title string, limit, offset int) ([]domain.Course, error)
	CountInstructorCourses(ctx context.Context, instructorID uuid.UUID, title string) (int, error)
	EnrolledStudents(ctx context.Context, courseID uuid.UUID, name string, limit, offset int) ([]domain.EnrolledStudent, error)
	CountEnrolledStudents(ctx context.Context, courseID uuid.UUID, name string) (int, error)
}

type store struct {
	q *db.Queries
}

func NewStore(q *db.Queries) Store {
	return &store{q: q}
}

func (s *store) AdminStats(ctx context.Context) (domain.AdminStats, error) {
	row, err := s.q.GetAdminStats(ctx)
	if err != nil {
		return domain.AdminStats{}, err
	}
	return domain.AdminStats{
		Users:       int(row.Users),
		Students:    int(row.Students),
		Instructors: int(row.Instructors),
		Courses:     int(row.Courses),
		Enrollments: int(row.Enrollments),
		Assignments: int(row.Assignments),
		Submissions: int(row.Submissions),
	}, nil
}

func (s *store) ListUsers(ctx context.Context, search, role string, limit, offset int) ([]domain.User, error) {
	rows, err := s.q.ListUsers(ctx, db.ListUsersParams{
		Search:      search,
		Role:        role,
		LimitCount:  int32(limit),
		OffsetCount: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		user, err := toDomainUser(row)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *store) CountUsers(ctx context.Context, search, role string) (int, error) {
	total, err := s.q.CountUsers(ctx, db.CountUsersParams{Search: search, Role: role})
	return int(total), err
}

func (s *store) InstructorStats(ctx context.Context, instructorID uuid.UUID) (domain.InstructorStats, error) {
	row, err := s.q.GetInstructorStats(ctx, db.PgUUID(instructorID))
	if err != nil {
		return domain.InstructorStats{}, err
	}
	return domain.InstructorStats{
		Courses:             int(row.Courses),
		PublishedCourses:    int(row.PublishedCourses),
		Enrollments:         int(row.Enrollments),
		Students:            int(row.Students),
		Lessons:             int(row.Lessons),
		Assignments:         int(row.Assignments),
		Submissions:         int(row.Submissions),
		UngradedSubmissions: int(row.UngradedSubmissions),
		Announcements:       int(row.Announcements),
	}, nil
}

func (s *store) StudentStats(ctx context.Context, studentID uuid.UUID) (domain.StudentStats, error) {
	row, err := s.q.GetStudentStats(ctx, db.PgUUID(studentID))
	if err != nil {
		return domain.StudentStats{}, err
	}
	return domain.StudentStats{
		EnrolledCourses:    int(row.EnrolledCourses),
		CompletedLessons:   int(row.CompletedLessons),
		Submissions:        int(row.Submissions),
		GradedSubmissions:  int(row.GradedSubmissions),
		AverageGrade:       row.AverageGrade,
		PendingAssignments: int(row.PendingAssignments),
	}, nil
}

func (s *store) StudentGrades(ctx context.Context, studentID uuid.UUID) (domain.StudentGradesDashboard, error) {
	rows, err := s.q.ListStudentGrades(ctx, db.PgUUID(studentID))
	if err != nil {
		return domain.StudentGradesDashboard{}, err
	}
	grades := make([]domain.StudentGrade, 0, len(rows))
	submitted := 0
	averageGrade := 0.0
	if len(rows) > 0 {
		submitted = int(rows[0].Submitted)
		averageGrade = rows[0].AverageGrade
	}
	for _, row := range rows {
		submissionID, err := db.FromPgUUID(row.SubmissionID)
		if err != nil {
			return domain.StudentGradesDashboard{}, err
		}
		assignmentID, err := db.FromPgUUID(row.AssignmentID)
		if err != nil {
			return domain.StudentGradesDashboard{}, err
		}
		courseID, err := db.FromPgUUID(row.CourseID)
		if err != nil {
			return domain.StudentGradesDashboard{}, err
		}
		var grade *int
		if row.Grade.Valid {
			value := int(row.Grade.Int32)
			grade = &value
		}
		grades = append(grades, domain.StudentGrade{
			SubmissionID:    submissionID,
			AssignmentID:    assignmentID,
			AssignmentTitle: row.AssignmentTitle,
			CourseID:        courseID,
			CourseTitle:     row.CourseTitle,
			Grade:           grade,
			Feedback:        row.Feedback.String,
			SubmittedAt:     row.SubmittedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		})
	}
	return domain.StudentGradesDashboard{
		Submitted:    submitted,
		AverageGrade: averageGrade,
		Grades:       grades,
	}, nil
}

func (s *store) CourseAnalytics(ctx context.Context, courseID uuid.UUID) (domain.CourseAnalytics, error) {
	row, err := s.q.GetCourseAnalytics(ctx, db.PgUUID(courseID))
	if err != nil {
		return domain.CourseAnalytics{}, err
	}
	return domain.CourseAnalytics{
		CourseID:          courseID,
		Enrollments:       int(row.Enrollments),
		Assignments:       int(row.Assignments),
		Submissions:       int(row.Submissions),
		AverageGrade:      row.AverageGrade,
		Lessons:           int(row.Lessons),
		LessonCompletions: int(row.LessonCompletions),
	}, nil
}

func (s *store) CourseProgress(ctx context.Context, studentID, courseID uuid.UUID) (domain.CourseProgress, error) {
	row, err := s.q.GetCourseProgress(ctx, db.GetCourseProgressParams{
		StudentID: db.PgUUID(studentID),
		CourseID:  db.PgUUID(courseID),
	})
	if err != nil {
		return domain.CourseProgress{}, err
	}
	total, completed := int(row.TotalLessons), int(row.CompletedLessons)
	percentage := 0.0
	if total > 0 {
		percentage = float64(completed) / float64(total) * 100
	}
	return domain.CourseProgress{
		CourseID:         courseID,
		TotalLessons:     total,
		CompletedLessons: completed,
		Percentage:       percentage,
	}, nil
}

func (s *store) Course(ctx context.Context, id uuid.UUID) (domain.Course, error) {
	row, err := s.q.GetCourseByID(ctx, db.PgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Course{}, errNotFound
		}
		return domain.Course{}, err
	}
	return toDomainCourse(courseListRow(row))
}

func (s *store) IsEnrolled(ctx context.Context, studentID, courseID uuid.UUID) (bool, error) {
	_, err := s.q.GetEnrollment(ctx, db.GetEnrollmentParams{
		UserID:   db.PgUUID(studentID),
		CourseID: db.PgUUID(courseID),
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func (s *store) Courses(ctx context.Context, published bool, title string, limit, offset int) ([]domain.Course, error) {
	rows, err := s.q.GetCourses(ctx, db.GetCoursesParams{
		Published:   published,
		Title:       title,
		LimitCount:  int32(limit),
		OffsetCount: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	return toDomainCourses(rows)
}

func (s *store) CountCourses(ctx context.Context, published bool, title string) (int, error) {
	total, err := s.q.CountCourses(ctx, db.CountCoursesParams{Published: published, Title: title})
	return int(total), err
}

func (s *store) EnrolledCourses(ctx context.Context, studentID uuid.UUID, title string, limit, offset int) ([]domain.Course, error) {
	rows, err := s.q.GetEnrolledCourses(ctx, db.GetEnrolledCoursesParams{
		StudentID:   db.PgUUID(studentID),
		Title:       title,
		LimitCount:  int32(limit),
		OffsetCount: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	return toDomainCourses(rows)
}

func (s *store) CountEnrolledCourses(ctx context.Context, studentID uuid.UUID, title string) (int, error) {
	total, err := s.q.CountEnrolledCourses(ctx, db.CountEnrolledCoursesParams{
		StudentID: db.PgUUID(studentID),
		Title:     title,
	})
	return int(total), err
}

func (s *store) InstructorCourses(ctx context.Context, instructorID uuid.UUID, title string, limit, offset int) ([]domain.Course, error) {
	rows, err := s.q.GetInstructorCourses(ctx, db.GetInstructorCoursesParams{
		InstructorID: db.PgUUID(instructorID),
		Title:        title,
		LimitCount:   int32(limit),
		OffsetCount:  int32(offset),
	})
	if err != nil {
		return nil, err
	}
	return toDomainCourses(rows)
}

func (s *store) CountInstructorCourses(ctx context.Context, instructorID uuid.UUID, title string) (int, error) {
	total, err := s.q.CountInstructorCourses(ctx, db.CountInstructorCoursesParams{
		InstructorID: db.PgUUID(instructorID),
		Title:        title,
	})
	return int(total), err
}

func (s *store) EnrolledStudents(ctx context.Context, courseID uuid.UUID, name string, limit, offset int) ([]domain.EnrolledStudent, error) {
	rows, err := s.q.ListEnrolledStudentsByCourse(ctx, db.ListEnrolledStudentsByCourseParams{
		CourseID:    db.PgUUID(courseID),
		Name:        name,
		LimitCount:  int32(limit),
		OffsetCount: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.EnrolledStudent, 0, len(rows))
	for _, row := range rows {
		id, err := db.FromPgUUID(row.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, domain.EnrolledStudent{
			ID:         id,
			Email:      row.Email,
			FullName:   row.FullName,
			EnrolledAt: row.EnrolledAt.Time,
		})
	}
	return out, nil
}

func (s *store) CountEnrolledStudents(ctx context.Context, courseID uuid.UUID, name string) (int, error) {
	total, err := s.q.CountEnrolledStudentsByCourse(ctx, db.CountEnrolledStudentsByCourseParams{
		CourseID: db.PgUUID(courseID),
		Name:     name,
	})
	return int(total), err
}
