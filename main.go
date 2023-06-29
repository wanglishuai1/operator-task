package main

import (
	"github.com/shenyisyn/myci/pkg/builder"
	"github.com/shenyisyn/myci/pkg/k8sconfig"
)

func main() {
	builder.InitImageCache(100)
	k8sconfig.InitManager()
}
