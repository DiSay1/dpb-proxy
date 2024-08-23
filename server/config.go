package server

type ListenerConfig struct {
	Address string
}

type RemoteConfig struct {
	Name    string
	Icon    string
	Address string
}

type GameConf struct {
	AuthenticationDisabled bool

	MaximumPlayers int
}

type ServerConfig struct {
	Listener ListenerConfig
	Servers  []RemoteConfig

	Game GameConf
}
