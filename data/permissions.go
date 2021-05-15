package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Permissions []string

type PermissionsModel struct {
	*sql.DB
}

func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

func (pm *PermissionsModel) GetAllForUser(id int64) (Permissions, error) {
	stmt := `SELECT permissions.code
	FROM permissions
	INNER JOIN user_permissions ON user_permissions.permission_id = permissions.id
	INNER JOIN users ON user_permissions.user_id = users.id
	WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	rows, err := pm.DB.QueryContext(ctx, stmt, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perm := Permissions{}
	for rows.Next() {
		var p string
		err = rows.Scan(&p)
		if err != nil {
			return nil, err
		}
		perm = append(perm, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perm, nil
}

func (pm *PermissionsModel) AddForUser(id int64, perms ...string) error {
	stmt := `INSERT INTO user_permissions
	SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`
	args := []interface{}{id, pq.Array(perms)}

	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeOut*time.Second)
	defer cancel()

	_, err := pm.DB.ExecContext(ctx, stmt, args...)
	return err
}
