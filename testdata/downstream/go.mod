module example.com/gg-yaml-downstream

go 1.26

require (
	github.com/mzz2017/gg v0.0.0
	gopkg.in/yaml.v3 v3.0.1
)

replace gopkg.in/yaml.v3 => ./legacyyaml
