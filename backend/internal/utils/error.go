package utils

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

func MapError(err error) map[string]string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errs := make(map[string]string)

		for _, fe := range ve {
			switch fe.Field() {
			case "Email":
				switch fe.Tag() {
				case "required":
					errs["email"] = "Email is required."
				case "email":
					errs["email"] = "Please enter a valid email address."
				}
			case "Password":
				switch fe.Tag() {
				case "required":
					errs["password"] = "Password is required."
				case "min":
					errs["password"] = "Password must be at least 8 characters long."
				}
			case "ConfirmPassword":
				switch fe.Tag() {
				case "required":
					errs["confirm_password"] = "Please confirm your password."
				case "eqfield":
					errs["confirm_password"] = "Passwords do not match."
				}

			case "OTP":
				switch fe.Tag() {
				case "required":
					errs["otp"] = "OTP is required."
				case "numeric":
					errs["otp"] = "OTP must be a number."
				}
			case "FullName":
				if fe.Tag() == "required" {
					errs["full_name"] = "Full name is required."
				}
			case "CourseID":
				if fe.Tag() == "required" {
					errs["course_id"] = "Course ID is required."
				}
			case "Title":
				if fe.Tag() == "required" {
					errs["title"] = "Title is required."
				}
			case "Description":
				if fe.Tag() == "required" {
					errs["description"] = "Description is required."
				}
			case "DueDate":
				if fe.Tag() == "required" {
					errs["due_date"] = "Due date is required."
				}
			case "CurrentPassword":
				if fe.Tag() == "required" {
					errs["current_password"] = "Current password is required."
				}

			}
		}
		return errs
	}
	return map[string]string{"error": err.Error()}
}
