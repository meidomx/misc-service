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
    time_created           bigint       NOT NULL,
    time_updated           bigint       NOT NULL,
    CONSTRAINT misc_ldap_entries_pkey PRIMARY KEY (entry_id)
);

create unique index if not exists misc_ldap_entries_uk
    ON misc_ldap_entries (entry_name, entry_type, parent_full_entry_path);

CREATE INDEX misc_ldap_entries_attr ON misc_ldap_entries USING gin (attribute);
CREATE INDEX misc_ldap_entries_meta ON misc_ldap_entries USING gin (metadata);

create table misc_ldap_uniqueness
(
    uniqueness_id    uuid,
    uniqueness_group varchar(300) NOT NULL,
    uniqueness_key   varchar(300) not null,
    entry_ref        uuid         NOT NULL,
    metadata         jsonb,
    time_created     bigint       NOT NULL,
    time_updated     bigint       NOT NULL,
    CONSTRAINT misc_ldap_uniqueness_pkey PRIMARY KEY (uniqueness_id)
);

create unique index if not exists misc_ldap_uniqueness_uk
    ON misc_ldap_uniqueness (uniqueness_group, uniqueness_key);

CREATE INDEX misc_ldap_uniqueness_meta ON misc_ldap_uniqueness USING gin (metadata);

