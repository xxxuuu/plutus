//go:generate abigen --abi=abi/PancakeFactoryV2.json --pkg book --type PancakeFactoryV2 --out generated_pancake_factory_v2.go
//go:generate abigen --abi=abi/PancakeRouterV2.json --pkg book --type PancakeRouterV2 --out generated_pancake_router_v2.go
//go:generate abigen --abi=abi/PancakePair.json --pkg book --type PancakePair --out generated_pancake_pair.go
//go:generate abigen --abi=abi/ERC20.json --pkg book --type Erc20 --out generated_erc20.go
package book
