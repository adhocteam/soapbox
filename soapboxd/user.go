package soapboxd

import (
	"crypto/hmac"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/adhocteam/soapbox/models"
	pb "github.com/adhocteam/soapbox/proto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

func (s *Server) CreateUser(ctx context.Context, user *pb.CreateUserRequest) (*pb.User, error) {
	password := []byte(user.Password)

	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	model := &models.User{
		ID:                int(user.Id),
		Name:              user.GetName(),
		Email:             user.GetEmail(),
		EncryptedPassword: string(hashedPassword),
	}

	if err := model.Insert(s.db); err != nil {
		return nil, errors.Wrap(err, "inserting into db")
	}

	user.Id = int32(model.ID)

	newUser := &pb.User{
		Id:                int32(model.ID),
		Name:              model.Name,
		Email:             model.Email,
		EncryptedPassword: model.EncryptedPassword,
	}

	return newUser, nil
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	if req.GetEmail() != "" {
		return s.getUserByEmail(ctx, req.GetEmail())
	}
	return s.getUserByID(ctx, int(req.GetId()))
}

func (s *Server) getUserByID(ctx context.Context, id int) (*pb.User, error) {
	model, err := models.UserByID(s.db, id)
	if err != nil {
		return nil, errors.Wrap(err, "getting user by id from db")
	}

	user := &pb.User{
		Id:                     int32(model.ID),
		Name:                   model.Name,
		Email:                  model.Email,
		EncryptedPassword:      model.EncryptedPassword,
		GithubOauthAccessToken: model.GithubOauthAccessToken,
	}

	return user, nil
}

func (s *Server) getUserByEmail(ctx context.Context, email string) (*pb.User, error) {
	model, err := models.UserByEmail(s.db, email)
	if err != nil {
		return nil, errors.Wrap(err, "getting user by email from db")
	}

	user := &pb.User{
		Id:                     int32(model.ID),
		Name:                   model.Name,
		Email:                  model.Email,
		EncryptedPassword:      model.EncryptedPassword,
		GithubOauthAccessToken: model.GithubOauthAccessToken,
	}

	return user, nil
}

func (s *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	const genericLoginErrorMsg = "could not log in user"

	res := &pb.LoginUserResponse{
		Error: genericLoginErrorMsg,
	}

	user, err := s.getUserByEmail(ctx, req.GetEmail())
	if err == sql.ErrNoRows {
		return res, nil
	} else if err != nil {
		return nil, errors.Cause(err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(req.Password)); err != nil {
		return res, nil
	}

	res.User = user
	res.Error = ""
	res.Hmac = computeHmac512(fmt.Sprint(user.Id), os.Getenv("LOGIN_SECRET_KEY"))

	return res, nil
}

func (s *Server) AssignGithubOmniauthTokenToUser(ctx context.Context, user *pb.User) (*pb.User, error) {
	model, err := models.UserByEmail(s.db, user.Email)
	if err != nil {
		return nil, errors.Wrap(err, "getting user by email from db")
	}

	model.GithubOauthAccessToken = user.GithubOauthAccessToken

	if err := model.Update(s.db); err != nil {
		return nil, errors.Wrap(err, "updating in db")
	}

	return user, nil
}

func computeHmac512(message string, secret string) string {
	h := hmac.New(sha512.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
