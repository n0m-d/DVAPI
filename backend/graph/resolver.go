package graph

import "github.com/n0m-d/DVAPI/internal/service"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	NoteService service.NoteService
}
