package local

type SystemOpts interface {
	customizeSystem(s *system)
}
