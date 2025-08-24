package main

import (
	"owenvi.com/fleetsim/internal/domainmodels"
)

func main() {

	domainmodels.PrintValGrid()
	loader := domainmodels.NewGridLoader()

	loader.ConfigureForTesting(20, 20, 99, 0.1, 0.05, 0.02, 0.9, 0.8, 0.15)
	print("NEXT")
	domainmodels.PrettyPrint(loader)
}
