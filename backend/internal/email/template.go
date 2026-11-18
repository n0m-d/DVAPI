package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*
var templateFS embed.FS

var bodyTpl *template.Template

func init() {
	var err error
	bodyTpl, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic("failed to parse email templates: " + err.Error())
	}
}

type otpEmailData struct {
	Name      string
	OTP       string
	CustomMsg string
}

type enrollmentEmailData struct {
	Name       string
	CourseName string
}

type welcomeEmailData struct {
	Name string
}

type gradeEmailData struct {
	Name            string
	CourseName      string
	AssignmentTitle string
	Grade           int
	Feedback        string
}

func sendOTPEmail(sender Sender, to, name, otp, customMsg string) error {
	var body bytes.Buffer
	if err := bodyTpl.ExecuteTemplate(&body, "otp_reset.html", otpEmailData{
		Name:      name,
		OTP:       otp,
		CustomMsg: customMsg,
	}); err != nil {
		return fmt.Errorf("render otp email body: %w", err)
	}
	return sender.Send(to, "Password Reset OTP", body.String())
}

func sendEnrollmentEmail(sender Sender, to, name, courseName string) error {
	var body bytes.Buffer
	if err := bodyTpl.ExecuteTemplate(&body, "enrollment.html", enrollmentEmailData{
		Name:       name,
		CourseName: courseName,
	}); err != nil {
		return fmt.Errorf("render enrollment email body: %w", err)
	}
	return sender.Send(to, "Course Enrollment", body.String())
}

func sendWelcomeEmail(sender Sender, to, name string) error {
	var body bytes.Buffer
	if err := bodyTpl.ExecuteTemplate(&body, "welcome.html", welcomeEmailData{Name: name}); err != nil {
		return fmt.Errorf("render welcome email body: %w", err)
	}
	return sender.Send(to, "Welcome to Schole", body.String())
}

func sendGradeEmail(sender Sender, to, name, courseName, assignmentTitle string, grade int, feedback string) error {
	var body bytes.Buffer
	if err := bodyTpl.ExecuteTemplate(&body, "grade.html", gradeEmailData{
		Name:            name,
		CourseName:      courseName,
		AssignmentTitle: assignmentTitle,
		Grade:           grade,
		Feedback:        feedback,
	}); err != nil {
		return fmt.Errorf("render grade email body: %w", err)
	}
	return sender.Send(to, "Grade Posted", body.String())
}
