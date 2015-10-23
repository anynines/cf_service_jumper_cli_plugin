package main

type ForwardSbCredentials struct {
	Uri             string `json:"uri"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	DefaultDatabase string `json:"default_database"`
	Database        string `json:"database"`
}

func (self ForwardSbCredentials) CredentialsMap() map[string]string {
	credentials := make(map[string]string)

	if len(self.Uri) > 0 {
		credentials["URI"] = self.Uri
	}
	if len(self.Username) > 0 {
		credentials["Username"] = self.Username
	}
	if len(self.Password) > 0 {
		credentials["Password"] = self.Password
	}
	if len(self.DefaultDatabase) > 0 {
		credentials["Default database"] = self.DefaultDatabase
	}
	if len(self.Database) > 0 {
		credentials["Database"] = self.Database
	}

	return credentials
}

type ForwardCredentials struct {
	Credentials ForwardSbCredentials `json:"credentials"`
}

type ForwardDataSet struct {
	ID           int                `json:"id"`
	Hosts        []string           `json:"public_uris"`
	SharedSecret string             `json:"shared_secret"`
	Credentials  ForwardCredentials `json:"credentials"`
}

// Returns map with credential information
func (self ForwardDataSet) CredentialsMap() map[string]string {
	return self.Credentials.Credentials.CredentialsMap()
}
