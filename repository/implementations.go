package repository

import "context"

func (r *Repository) GetTestById(ctx context.Context, input GetTestByIdInput) (output GetTestByIdOutput, err error) {
	err = r.Db.QueryRowContext(ctx, "SELECT name FROM test WHERE id = $1", input.Id).Scan(&output.Name)
	if err != nil {
		return
	}
	return
}

func (r *Repository) Registration(ctx context.Context, input RegistrationInput) (output RegistrationOutput, err error) {
	_, err = r.Db.ExecContext(ctx, "INSERT INTO public.user (id, phone, name, password, salt) VALUES ($1, $2, $3, $4, $5)", input.ID, input.PhoneNumber, input.FullName, input.Password, input.Salt)
	if err != nil {
		return
	}
	return RegistrationOutput{ID: input.ID}, nil
}
