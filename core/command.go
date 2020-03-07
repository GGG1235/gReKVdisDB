package core

type GkvDBCommand struct {
	Name string
	Proc cmdFunc
}

type cmdFunc func(c *Client, s *Server)

func lookupCommand(name string, s *Server) *GkvDBCommand {
	if cmd, ok := s.Commands[name]; ok {
		return cmd
	}
	return nil
}