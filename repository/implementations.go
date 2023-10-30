package repository

import (
	"context"
	"fmt"
	"log"
	"strconv"
)

func (r *Repository) GetTestById(ctx context.Context, input GetTestByIdInput) (output GetTestByIdOutput, err error) {
	err = r.Db.QueryRowContext(ctx, "SELECT name FROM test WHERE id = $1", input.Id).Scan(&output.Name)
	if err != nil {
		return
	}
	return
}

func (r *Repository) Registration(ctx context.Context, input RegistrationInput) (output RegistrationOutput, err error) {
	_, err = r.Db.ExecContext(ctx, "INSERT INTO public.user (id, phone, name, password, salt) VALUES ($1, $2, $3, $4, $5)", input.ID, input.Phone, input.Name, input.Password, input.Salt)
	if err != nil {
		return
	}
	return RegistrationOutput{ID: input.ID}, nil
}

// FindUser : Find user by params
func (r *Repository) FindUser(ctx context.Context, params ...Param) (user User, err error) {
	where := ""
	var values []any
	for i, param := range params {
		if where != "" {
			logic := "AND "
			if param.Logic != "" {
				logic = param.Logic + " "
			}
			where += logic
		}
		where += param.Field + " " + param.Operator + " $" + strconv.Itoa(i+1) + ""
		values = append(values, param.Value)
	}

	if where != "" {
		where = "WHERE " + where
	}
	log.Println(fmt.Sprintf("SELECT id, phone, name, password, salt FROM public.user %s %v", where, values))
	err = r.Db.QueryRowContext(ctx, fmt.Sprintf("SELECT id, phone, name, password, salt FROM public.user %s", where), values...).Scan(&user.ID, &user.Phone, &user.Name, &user.Password, &user.Salt)
	if err != nil {
		return
	}
	return
}

func (r *Repository) IncreaseLoginAttempt(ctx context.Context, phone string) (err error) {
	_, err = r.Db.ExecContext(ctx, fmt.Sprintf("UPDATE public.user SET success_login = (SELECT success_login FROM public.user WHERE phone = $1) + 1, updated_at=NOW() WHERE phone = $2"), phone, phone)
	if err != nil {
		return
	}
	return
}

func (r *Repository) UpdateUser(ctx context.Context, user UpdateUser) (err error) {
	_, err = r.Db.ExecContext(ctx, "UPDATE public.user SET phone=$1, name=$2, updated_at=NOW() WHERE id=$3", user.Phone, user.Name, user.ID)
	if err != nil {
		return
	}
	return
}
