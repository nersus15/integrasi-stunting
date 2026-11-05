package testutil

import "database/sql"

// key API tetap untuk seluruh test, dibuat EnsureAPIKeys kalau belum ada
const (
	KeyJakantro = "key-jakantro"
	KeyFaskesA  = "key-faskes-A"
	KeyFaskesB  = "key-faskes-B"
	KeyReadonly = "key-readonly"

	OrgFaskesA = "100011"
	OrgFaskesB = "100009864"
)

// satu user per key karena auth_apikey.user_id unik
func EnsureAPIKeys(db *sql.DB) error {
	_, err := db.Exec(`
WITH u AS (
  INSERT INTO access.auth_users (username, password_hash, groups, roles) VALUES
    ('uji-key-jakantro', '-', '{service,jakantro}', '{read,write}'),
    ('uji-key-faskes-a', '-', '{orgid:`+OrgFaskesA+`}', '{read,write}'),
    ('uji-key-faskes-b', '-', '{orgid:`+OrgFaskesB+`}', '{read,write}'),
    ('uji-key-readonly', '-', '{orgid:`+OrgFaskesA+`}', '{read}')
  ON CONFLICT (username) DO UPDATE SET groups = EXCLUDED.groups, roles = EXCLUDED.roles, active = true
  RETURNING id, username)
INSERT INTO access.auth_apikey (api_key, user_id)
SELECT CASE username
  WHEN 'uji-key-jakantro' THEN $1 WHEN 'uji-key-faskes-a' THEN $2
  WHEN 'uji-key-faskes-b' THEN $3 ELSE $4 END, id FROM u
ON CONFLICT (user_id) DO UPDATE SET api_key = EXCLUDED.api_key`, KeyJakantro, KeyFaskesA, KeyFaskesB, KeyReadonly)
	return err
}
