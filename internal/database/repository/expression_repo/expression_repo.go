package expressionrepo

import (
	"database/sql"

	"github.com/MoodyShoo/go-http-calculator/internal/models"
)

type ExpressionRepo struct {
	Db *sql.DB
}

func (er *ExpressionRepo) InsertExpression(exp models.Expression) error {
	query := `INSERT INTO expressions (expression, status, result, error, user_id)
				VALUES ($1, $2, $3, $4, $5)`

	_, err := er.Db.Exec(query, exp.Expr, exp.Status, exp.Result, exp.Error, exp.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (er *ExpressionRepo) UpdateExpression(id int64, newExpr models.Expression) error {
	query := `UPDATE expressions 
			  SET status = $1, result = $2, error = $3 
			  WHERE id = $4`

	_, err := er.Db.Exec(query, newExpr.Status, newExpr.Result, newExpr.Error, id)
	if err != nil {
		return err
	}

	return nil
}

func (er *ExpressionRepo) GetExpressionByID(id int64) (models.Expression, error) {
	e := models.Expression{}
	query := `SELECT * FROM expressions WHERE id = $1`

	err := er.Db.QueryRow(query, id).Scan(&e.Id, &e.Expr, &e.Status, &e.Result, &e.Error, &e.UserID)
	if err != nil {
		return models.Expression{}, err
	}

	return e, nil
}

func (er *ExpressionRepo) GetExpressions() ([]models.Expression, error) {
	var expressions []models.Expression
	query := "SELECT * FROM expressions"

	rows, err := er.Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := models.Expression{}
		err := rows.Scan(&e.Id, &e.Expr, &e.Status, &e.Result, &e.Error, &e.UserID)
		if err != nil {
			return nil, err
		}

		expressions = append(expressions, e)
	}

	return expressions, nil
}

func (er *ExpressionRepo) GetComputingAndPending() ([]models.Expression, error) {
	var expressions []models.Expression
	query := `SELECT * FROM expressions WHERE status = $1 OR status = $2`

	rows, err := er.Db.Query(query, models.StatusComputing, models.StatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := models.Expression{}
		err := rows.Scan(&e.Id, &e.Expr, &e.Status, &e.Result, &e.Error, &e.UserID)
		if err != nil {
			return nil, err
		}

		expressions = append(expressions, e)
	}

	return expressions, nil
}
