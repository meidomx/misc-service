package config

type Config struct {
	Storage struct {
		ConnString string `toml:"conn_string"`
	} `toml:"storage"`

	Http struct {
		Address string `toml:"address"`
	} `toml:"http"`

	LDAP struct {
		Address    string `toml:"address"`
		BindBaseDN string `toml:"bind_base_dn"`

		TLS struct {
			Enable     bool   `toml:"enable"`
			TLSAddress string `toml:"tls_address"`
			CertPath   string `toml:"cert_path"`
			KeyPath    string `toml:"key_path"`
		} `toml:"tls"`

		Init struct {
			InitScripts []struct {
				DN            string   `toml:"dn"`
				ObjectClasses []string `toml:"object_classes"`
			} `toml:"run_simple_init_scripts"`
			InitAdmin struct {
				DN            string   `toml:"dn"`
				ObjectClasses []string `toml:"object_classes"`
				UserPassword  string   `toml:"user_password"`
			} `toml:"run_init_admin"`
		} `toml:"init"`
	} `toml:"ldap"`

	FullTextSearch struct {
		IndexFolder string `toml:"index_folder"`
	} `toml:"full_text_search"`
}
