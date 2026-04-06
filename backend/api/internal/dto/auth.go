package dto

import "github.com/google/uuid"

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r RegisterRequest) Validate() map[string]string {
	errs := map[string]string{}
	if len(r.Username) < 3 || len(r.Username) > 50 {
		errs["username"] = "must be between 3 and 50 characters"
	}
	if len(r.Email) < 5 || len(r.Email) > 255 {
		errs["email"] = "must be a valid email"
	}
	if len(r.Password) < 8 || len(r.Password) > 128 {
		errs["password"] = "must be between 8 and 128 characters"
	}
	return errs
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() map[string]string {
	errs := map[string]string{}
	if r.Email == "" {
		errs["email"] = "required"
	}
	if r.Password == "" {
		errs["password"] = "required"
	}
	return errs
}

type AuthResponse struct {
	AccessToken string   `json:"access_token"`
	User        UserDTO  `json:"user"`
}

type UserDTO struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}
