package ldap

import (
	"context"
	"time"

	"github.com/meidomx/misc-service/id"
	"github.com/meidomx/misc-service/pgbackend"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jimlambrt/gldap"
)

const (
	EntryTypeDomainComponent  = "dc"
	EntryTypeCountry          = "c"
	EntryTypeOrganization     = "o"
	EntryTypeOrganizationUnit = "ou"
	EntryTypeSurname          = "sn"
	EntryTypeCommonName       = "cn"
)

const (
	ServiceName = "ldap"
)

func FindSingleRoot() (*gldap.Entry, error) {
	entry := new(gldap.Entry)
	r, err := pgbackend.RunQuery(ServiceName, entry, func(conn *pgxpool.Conn, result *gldap.Entry) error {
		rows, err := conn.Query(context.Background(),
			"select attribute from misc_ldap_entries where parent_full_entry_path IS NULL LIMIT 1",
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			if err := rows.Scan(result); err != nil {
				return err
			}
		}

		return nil
	})

	return r, err
}

func FindAllRoots() ([]*gldap.Entry, error) {
	var entries []*gldap.Entry

	r, err := pgbackend.RunQuery(ServiceName, &entries, func(conn *pgxpool.Conn, r *[]*gldap.Entry) error {
		rows, err := conn.Query(context.Background(),
			"select attribute from misc_ldap_entries where parent_full_entry_path IS NULL",
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		var rr []*gldap.Entry
		for rows.Next() {
			result := new(gldap.Entry)
			if err := rows.Scan(result); err != nil {
				return err
			}
			rr = append(rr, result)
		}
		*r = rr

		return nil
	})

	return *r, err
}

func FindOneEntry(dn string) (*gldap.Entry, error) {
	entryPath := SplitDN(dn)
	entry := new(gldap.Entry)
	r, err := pgbackend.RunQuery(ServiceName, entry, func(conn *pgxpool.Conn, result *gldap.Entry) error {
		parent := CombineParentDN(entryPath)
		if len(parent) != 0 {
			rows, err := conn.Query(context.Background(),
				"select attribute from misc_ldap_entries where entry_name = $1 and parent_full_entry_path = $2",
				entryPath[0], parent)
			if err != nil {
				return err
			}
			defer rows.Close()

			if rows.Next() {
				if err := rows.Scan(result); err != nil {
					return err
				}
			}

			return nil
		} else {
			rows, err := conn.Query(context.Background(),
				"select attribute from misc_ldap_entries where entry_name = $1 and parent_full_entry_path is NULL",
				entryPath[0])
			if err != nil {
				return err
			}
			defer rows.Close()

			if rows.Next() {
				if err := rows.Scan(result); err != nil {
					return err
				}
			}

			return nil
		}
	})

	return r, err
}

func FindChildren(dn string) ([]*gldap.Entry, error) {

	entryPath := SplitDN(dn)
	var entries []*gldap.Entry
	r, err := pgbackend.RunQuery(ServiceName, &entries, func(conn *pgxpool.Conn, r *[]*gldap.Entry) error {
		parent := CombineDN(entryPath)
		rows, err := conn.Query(context.Background(),
			"select attribute from misc_ldap_entries where parent_full_entry_path = $1",
			parent)
		if err != nil {
			return err
		}
		defer rows.Close()

		var rr []*gldap.Entry
		for rows.Next() {
			result := new(gldap.Entry)
			if err := rows.Scan(result); err != nil {
				return err
			}
			rr = append(rr, result)
		}
		*r = rr

		return nil
	})

	return *r, err
}

func SaveEntry(entry *gldap.Entry, i id.ItemId) error {
	_, err := pgbackend.RunQuery(ServiceName, entry, func(conn *pgxpool.Conn, result *gldap.Entry) error {
		sp := SplitDN(entry.DN)
		entryType := EntryType(sp[0])
		parent := CombineParentDN(sp)
		now := time.Now().UnixMilli()
		if len(parent) > 0 {
			_, err := conn.Exec(context.Background(),
				"insert into misc_ldap_entries(entry_id, entry_name, parent_full_entry_path, entry_type, attribute, metadata, time_created, time_updated) values($1, $2, $3, $4, $5, $6, $7, $8)",
				i.HexString(), sp[0], parent, entryType, result, nil, now, now)
			return err
		} else {
			_, err := conn.Exec(context.Background(),
				"insert into misc_ldap_entries(entry_id, entry_name, entry_type, attribute, metadata, time_created, time_updated) values($1, $2, $3, $4, $5, $6, $7)",
				i.HexString(), sp[0], entryType, result, nil, now, now)
			return err
		}
	})
	return err
}

func UpdateEntry(entry *gldap.Entry) error {
	_, err := pgbackend.RunQuery(ServiceName, entry, func(conn *pgxpool.Conn, result *gldap.Entry) error {
		sp := SplitDN(entry.DN)
		entryType := EntryType(sp[0])
		parent := CombineParentDN(sp)
		now := time.Now().UnixMilli()
		if len(parent) > 0 {
			_, err := conn.Exec(context.Background(),
				"update misc_ldap_entries set attribute = $1, time_updated = $2 where entry_name = $3 and parent_full_entry_path = $4 and entry_type = $5",
				entry, now, sp[0], parent, entryType)
			return err
		} else {
			_, err := conn.Exec(context.Background(),
				"update misc_ldap_entries set attribute = $1, time_updated = $2 where entry_name = $3 and parent_full_entry_path IS NULL and entry_type = $4",
				entry, now, sp[0], entryType)
			return err
		}
	})
	return err
}

func DeleteEntry(dn string) error {
	_, err := pgbackend.RunQuery(ServiceName, nil, func(conn *pgxpool.Conn, result interface{}) error {
		sp := SplitDN(dn)
		entryType := EntryType(sp[0])
		parent := CombineParentDN(sp)
		if len(parent) > 0 {
			_, err := conn.Exec(context.Background(),
				"delete from misc_ldap_entries where entry_name = $1 and entry_type = $2 and parent_full_entry_path = $3",
				sp[0], entryType, parent)
			return err
		} else {
			_, err := conn.Exec(context.Background(),
				"delete from misc_ldap_entries where entry_name = $1 and entry_type = $2 and parent_full_entry_path IS NULL",
				sp[0], entryType)
			return err
		}
	})

	return err
}
