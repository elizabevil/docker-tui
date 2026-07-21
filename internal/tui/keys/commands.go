package keys

const (
	CommandCompose    = "compose"
	CommandImages     = "images"
	CommandContainers = "containers"
	CommandVolumes    = "volumes"
	CommandNetworks   = "networks"
	CommandLogs       = "logs"
	CommandHelp       = "help"
)

// Commands returns command-mode entries in autocomplete order.
func Commands() []string {
	return []string{CommandCompose, CommandImages, CommandContainers, CommandVolumes, CommandNetworks, CommandLogs, CommandHelp}
}
