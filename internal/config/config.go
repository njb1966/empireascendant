package config

const (
	GameID                 = "empire_ascendant"
	GameTitle              = "Empire Ascendant"
	GameVersion            = "0.4.0"
	DefaultProtocolVersion = "1"
	DefaultAddr            = ":2324"
	DefaultDB              = "var/empireascendant.db"
	DefaultSSHHostKey      = "var/ssh_host_ed25519"
	DefaultANSI            = true
	DefaultSSHEncoding     = "auto"
)

type Config struct {
	DBPath            string
	Addr              string
	Stdio             bool
	ANSI              bool
	SSHEncoding       string
	SSHHostKeyPath    string
	NodeID            string
	HubURL            string
	RegistrationToken string
	APIKey            string
	AdvertiseAddr     string
	GameID            string
	GameTitle         string
	GameVersion       string
	ProtocolVersion   string
}

func Default() Config {
	return Config{
		DBPath:          DefaultDB,
		Addr:            DefaultAddr,
		Stdio:           true,
		ANSI:            DefaultANSI,
		SSHEncoding:     DefaultSSHEncoding,
		SSHHostKeyPath:  DefaultSSHHostKey,
		GameID:          GameID,
		GameTitle:       GameTitle,
		GameVersion:     GameVersion,
		ProtocolVersion: DefaultProtocolVersion,
	}
}
