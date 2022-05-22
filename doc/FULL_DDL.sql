------------------------------------------------------------------------
-- LDAP tables
------------------------------------------------------------------------
create table misc_ldap_entries
(
    entry_id               uuid,
    entry_name             varchar(300) NOT NULL,
    parent_full_entry_path varchar(300),
    entry_type             varchar(200) NOT NULL,
    attribute              jsonb,
    metadata               jsonb,
    CONSTRAINT misc_ldap_entries_pkey PRIMARY KEY (entry_id)
);

create unique index if not exists misc_ldap_entries_uk
    ON misc_ldap_entries (parent_full_entry_path, entry_name, entry_type);

CREATE INDEX misc_ldap_entries_attr ON misc_ldap_entries USING gin (attribute);
CREATE INDEX misc_ldap_entries_attr ON misc_ldap_entries USING gin (metadata);

