package runtime

const basePath = "/var/run/forker"

type Sandbox struct {
	ID  string
	PID int
	Cmd string
}
