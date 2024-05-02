package auth_handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nahtann/go-lab/internal/utils"
)

type SignUpHttpHandler struct {
	DB *pgxpool.Pool
}

type SignupRequest struct {
	Username string `json:"username" validate:"required" example:"NahtanN"`
	Email    string `json:"email"    validate:"required" example:"nahtann@outlook.com"`
	Password string `json:"password" validate:"required" example:"#Asdf123"`
}

//	@Description	Creates new user.
//	@Tags			auth
//	@Accept			json
//	@Param			request	body	SignupRequest	true	"Request Body"
//	@Produce		json
//	@Success		201	{object}	utils.DefaultResponse	"Message: 'Sign up successfully'"
//	@Failure		400	{object}	utils.CustomError		"Message: 'Username already in use. E-mail already in use.'"
//	@router			/auth/sign-up   [post]
func NewSignUpHttpHandler(db *pgxpool.Pool) *SignUpHttpHandler {
	return &SignUpHttpHandler{
		DB: db,
	}
}

func (handler *SignUpHttpHandler) Serve(w http.ResponseWriter, r *http.Request) error {
	request := new(SignupRequest)

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return utils.HttpServerInvalidRequest(w)
	}

	errorMessages := utils.Validate(request)
	if errorMessages != "" {
		message := utils.DefaultResponse{
			Message: errorMessages,
		}

		return utils.WriteJSON(w, http.StatusBadRequest, message)
	}

	err = SingUp(handler.DB, request)
	if err != nil {
		return utils.WriteJSON(w, http.StatusBadRequest, err)
	}

	message := utils.DefaultResponse{
		Message: "Sign up successfully",
	}

	return utils.WriteJSON(w, http.StatusCreated, message)
}

func SingUp(db *pgxpool.Pool, request *SignupRequest) error {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(%s) LIKE LOWER($1))"

	errorMessages := []string{}

	var usernameExists, emailExists bool

	queryByUsername := fmt.Sprintf(query, "username")

	err := db.QueryRow(context.Background(), queryByUsername, request.Username).
		Scan(&usernameExists)
	if err != nil {
		return &utils.CustomError{
			Message: "Unable to validate user username.",
		}
	}
	if usernameExists {
		errorMessages = append(errorMessages, "Username already in use.")
	}

	queryByEmail := fmt.Sprintf(query, "email")

	err = db.QueryRow(context.Background(), queryByEmail, request.Email).
		Scan(&emailExists)
	if err != nil {
		return &utils.CustomError{
			Message: "Unable to validate user email.",
		}
	}
	if emailExists {
		errorMessages = append(errorMessages, "E-mail already in use.")
	}

	if usernameExists || emailExists {
		return &utils.CustomError{
			Message: strings.Join(errorMessages, " "),
		}
	}

	hashPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		return &utils.CustomError{
			Message: "Unable to validate password.",
		}
	}

	_, err = db.Exec(
		context.Background(),
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		request.Username, request.Email, hashPassword,
	)
	if err != nil {
		return &utils.CustomError{
			Message: "Unable to create user.",
		}
	}

	return nil
}
