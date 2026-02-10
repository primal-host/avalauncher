package docker

import "github.com/docker/docker/api/types/filters"

func newFilterArgs(key, value string) filters.Args {
	f := filters.NewArgs()
	f.Add("label", key+"="+value)
	return f
}
