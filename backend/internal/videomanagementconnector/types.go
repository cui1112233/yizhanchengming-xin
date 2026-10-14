package videomanagementconnector

import "context"

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Session struct {
	Cookie         string `json:"cookie"`
	TargetUsername string `json:"targetUsername,omitempty"`
}

type CatalogItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Catalogs struct {
	Organizations []CatalogItem `json:"organizations,omitempty"`
	Raw           any           `json:"raw,omitempty"`
}

type SubmissionInput struct {
	BookID     string            `json:"bookId"`
	Filename   string            `json:"filename"`
	Content    []byte            `json:"content"`
	FormFields map[string]string `json:"formFields,omitempty"`
}

type Submission struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	ReadbackRaw any    `json:"readbackRaw,omitempty"`
}

type Service interface {
	Login(context.Context, LoginInput) (Session, error)
	RefreshCatalogs(context.Context, Session) (Catalogs, error)
	Submit(context.Context, Session, SubmissionInput) (Submission, error)
	GetSubmission(context.Context, Session, string) (Submission, error)
}
