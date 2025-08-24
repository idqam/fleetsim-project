package main

import (
	"owenvi.com/fleetsim/internal/domainmodels"
)

func main() {


	domainmodels.PrintValGrid()
	loader := domainmodels.NewGridLoader()
	domainmodels.PrettyPrint(loader)

	
	loader.ConfigureForTesting(10, 10, 99, 0.1, 0.05, 0.02, 0.6, 0.25, 0.15)
	print("NEXT")
	domainmodels.PrettyPrint(loader)
}
