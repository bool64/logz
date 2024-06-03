module zapage

go 1.17

replace github.com/bool64/logz => ./../..

require (
	github.com/bool64/logz v0.0.0-00010101000000-000000000000
	github.com/drhodes/golorem v0.0.0-20160418191928-ecccc744c2d9
	go.uber.org/zap v1.27.0
)

require (
	github.com/vearutop/dynhist-go v1.2.3 // indirect
	github.com/vearutop/lograte v1.1.3 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
)
