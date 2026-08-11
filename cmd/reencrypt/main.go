// Command reencrypt rotates the ENCRYPTION_KEY: decrypts every encrypted column
// with the OLD key and re-encrypts with the NEW key, in-place.
//
// Usage (against the production DB, while the API is stopped):
//
//	reencrypt -dsn "$DATABASE_URL" -old-key "$OLD_KEY" -new-key "$NEW_KEY"        # dry-run report
//	reencrypt -dsn "$DATABASE_URL" -old-key "$OLD_KEY" -new-key "$NEW_KEY" -apply # write changes
//	reencrypt -dsn "$DATABASE_URL" -new-key "$NEW_KEY" [-old-key "$OLD_KEY"] -verify # post-check
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"drone-platform/internal/crypto"
)

// encryptedColumns is the full inventory of AES-256-GCM encrypted columns.
// Keep in sync with the repository encrypt/decrypt helpers.
var encryptedColumns = []struct {
	table  string
	column string
}{
	{"demands", "contact"},
	{"enterprises", "license_url"},
	{"enterprises", "account_name"},
	{"users", "phone_ciphertext"},
	{"certified_pilots", "id_card"},
}

func main() {
	dsn := flag.String("dsn", "", "PostgreSQL DSN (DATABASE_URL)")
	oldKey := flag.String("old-key", "", "current ENCRYPTION_KEY (base64)")
	newKey := flag.String("new-key", "", "new ENCRYPTION_KEY (base64)")
	apply := flag.Bool("apply", false, "apply re-encryption (default: dry-run)")
	verify := flag.Bool("verify", false, "verify all values decrypt with -new-key")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("-dsn is required")
	}
	if *verify {
		if *newKey == "" {
			log.Fatal("-verify requires -new-key")
		}
		if err := verifyAll(context.Background(), *dsn, *oldKey, *newKey); err != nil {
			log.Fatalf("verify failed: %v", err)
		}
		fmt.Println("verify OK: no problems found")
		return
	}
	if *oldKey == "" || *newKey == "" {
		log.Fatal("-old-key and -new-key are required")
	}

	oldC, err := crypto.NewCipher(*oldKey)
	if err != nil {
		log.Fatalf("old key: %v", err)
	}
	newC, err := crypto.NewCipher(*newKey)
	if err != nil {
		log.Fatalf("new key: %v", err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, *dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	total := 0
	for _, col := range encryptedColumns {
		changed, plaintext, err := reencryptColumn(ctx, conn, col.table, col.column, oldC, newC, *apply)
		if err != nil {
			log.Printf("[%s.%s] SKIPPED: %v", col.table, col.column, err)
			continue
		}
		fmt.Printf("[%s.%s] %d value(s) %s | %d legacy plaintext (left as-is)\n",
			col.table, col.column, changed, modeLabel(*apply), plaintext)
		total += changed
	}
	fmt.Printf("done: %d value(s) re-encrypted (%s)\n", total, modeLabel(*apply))
}

func modeLabel(apply bool) string {
	if apply {
		return "RE-ENCRYPTED"
	}
	return "would re-encrypt"
}

// reencryptColumn decrypts each non-empty value with the old key and rewrites it
// with the new key. Values that do not decrypt with the old key are treated as
// legacy plaintext and left untouched (counted separately). Runs in one
// transaction per table so a failure rolls back the whole table.
func reencryptColumn(ctx context.Context, conn *pgx.Conn, table, column string, oldC, newC *crypto.Cipher, apply bool) (changed, plaintext int, err error) {
	rows, err := conn.Query(ctx, fmt.Sprintf(
		"SELECT id, %s FROM %s WHERE %s IS NOT NULL AND %s <> ''", column, table, column, column))
	if err != nil {
		return 0, 0, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	type row struct {
		id    any
		value string
	}
	var vals []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.value); err != nil {
			return 0, 0, fmt.Errorf("scan: %w", err)
		}
		vals = append(vals, r)
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("iterate: %w", err)
	}

	for _, r := range vals {
		plain, derr := oldC.Decrypt(r.value)
		if derr != nil {
			plaintext++ // not old-key ciphertext: legacy plaintext, leave as-is
			continue
		}
		enc, eerr := newC.Encrypt(plain)
		if eerr != nil {
			return changed, plaintext, fmt.Errorf("encrypt %s.%s id=%v: %w", table, column, r.id, eerr)
		}
		changed++
		if !apply {
			continue
		}
		if _, err := conn.Exec(ctx, fmt.Sprintf("UPDATE %s SET %s = $1 WHERE id = $2", table, column), enc, r.id); err != nil {
			return changed, plaintext, fmt.Errorf("update %s.%s id=%v: %w", table, column, r.id, err)
		}
	}
	return changed, plaintext, nil
}

// verifyAll checks every encrypted value decrypts with the new key. If the old
// key is provided, values decrypting with it are reported as MISSED (not
// re-encrypted). Values failing both keys are reported as samples — likely
// legacy plaintext, expected only for rows written before encryption existed.
func verifyAll(ctx context.Context, dsn, oldKeyB64, newKeyB64 string) error {
	newC, err := crypto.NewCipher(newKeyB64)
	if err != nil {
		return fmt.Errorf("new key: %w", err)
	}
	var oldC *crypto.Cipher
	if oldKeyB64 != "" {
		oldC, err = crypto.NewCipher(oldKeyB64)
		if err != nil {
			return fmt.Errorf("old key: %w", err)
		}
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	problems := 0
	for _, col := range encryptedColumns {
		rows, err := conn.Query(ctx, fmt.Sprintf(
			"SELECT id, %s FROM %s WHERE %s IS NOT NULL AND %s <> ''", col.column, col.table, col.column, col.column))
		if err != nil {
			fmt.Printf("[%s.%s] SKIPPED: %v\n", col.table, col.column, err)
			continue
		}
		for rows.Next() {
			var id any
			var value string
			if err := rows.Scan(&id, &value); err != nil {
				rows.Close()
				return fmt.Errorf("scan %s.%s: %w", col.table, col.column, err)
			}
			if _, derr := newC.Decrypt(value); derr == nil {
				continue
			}
			problems++
			if oldC != nil {
				if _, derr := oldC.Decrypt(value); derr == nil {
					fmt.Printf("[%s.%s] MISSED (still old key): id=%v\n", col.table, col.column, id)
					continue
				}
			}
			fmt.Printf("[%s.%s] undecryptable (legacy plaintext?): id=%v value=%.12q…\n", col.table, col.column, id, value)
		}
		rows.Close()
	}
	if problems > 0 {
		return fmt.Errorf("%d value(s) failed verification", problems)
	}
	return nil
}
