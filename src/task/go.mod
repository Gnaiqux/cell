module github.com/project-nano/cell/task

go 1.13

replace (
	github.com/project-nano/cell/service => ../service
	github.com/project-nano/framework => ../../../framework
)

require (
	github.com/pkg/errors v0.9.1
	github.com/project-nano/cell/service v0.0.0-00010101000000-000000000000
	github.com/project-nano/framework v1.0.9
)
